package push

import (
	"bytes"
	"chat-server/internals/db/models"
	"chat-server/internals/events"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/jwt"
)

type channelMembers interface {
	FindMembersByChannelID(context.Context, uuid.UUID) ([]models.ChannelMember, error)
}

type Sender struct {
	store    *Store
	channels channelMembers
	client   *http.Client
	url      string
	queue    chan notification
}

type notification struct {
	kind     string
	targetID uuid.UUID
	channel  *uuid.UUID
	actor    uuid.UUID
	users    []uuid.UUID
}

var projectPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,62}$`)

func NewSender(store *Store, channels channelMembers) (*Sender, error) {
	project, filename := os.Getenv("FCM_PROJECT_ID"), os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if project == "" {
		return nil, nil
	}
	if !projectPattern.MatchString(project) || filename == "" {
		return nil, fmt.Errorf("FCM needs a project ID and a service-account file")
	}
	raw, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("could not read the FCM service-account file")
	}
	var account struct {
		Email string `json:"client_email"`
		Key   string `json:"private_key"`
		Type  string `json:"type"`
	}
	if err := json.Unmarshal(raw, &account); err != nil || account.Type != "service_account" ||
		account.Email == "" || account.Key == "" {
		return nil, fmt.Errorf("invalid FCM service-account file")
	}
	base := &http.Client{Timeout: 8 * time.Second}
	authContext := context.WithValue(context.Background(), oauth2.HTTPClient, base)
	config := &jwt.Config{
		Email: account.Email, PrivateKey: []byte(account.Key),
		TokenURL: "https://oauth2.googleapis.com/token",
		Scopes:   []string{"https://www.googleapis.com/auth/firebase.messaging"},
	}
	client := oauth2.NewClient(authContext, config.TokenSource(authContext))
	client.Timeout = base.Timeout
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	return &Sender{
		store: store, channels: channels, client: client,
		url:   "https://fcm.googleapis.com/v1/projects/" + project + "/messages:send",
		queue: make(chan notification, 256),
	}, nil
}

func (s *Sender) Message(message events.MessageEvent) {
	if message.ChannelID != nil {
		s.enqueue(notification{kind: "channel", targetID: *message.ChannelID, channel: message.ChannelID, actor: message.SenderID})
	} else if message.ReceiverID != nil && *message.ReceiverID != message.SenderID {
		s.enqueue(notification{kind: "dm", targetID: message.SenderID, users: []uuid.UUID{*message.ReceiverID}})
	}
}

func (s *Sender) Plan(plan events.PlanEvent) {
	switch plan.Event {
	case events.PlanCreated, events.PlanUpdated, events.PlanConfirmed, events.PlanCancelled, events.PlanStatusChanged:
		s.enqueue(notification{kind: "plan", targetID: plan.PlanID, users: plan.Recipients, actor: plan.CreatorID})
	}
}

func (s *Sender) enqueue(n notification) {
	select {
	case s.queue <- n:
	default:
		// Push is best effort. The message or plan is already in the database.
	}
}

func (s *Sender) Run() {
	for n := range s.queue {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		s.deliver(ctx, n)
		cancel()
	}
}

func (s *Sender) deliver(ctx context.Context, n notification) {
	users := make([]uuid.UUID, 0, len(n.users))
	for _, userID := range n.users {
		if userID != n.actor {
			users = append(users, userID)
		}
	}
	if n.channel != nil {
		members, err := s.channels.FindMembersByChannelID(ctx, *n.channel)
		if err != nil {
			log.Print("push channel recipients unavailable")
			return
		}
		for _, member := range members {
			if member.UserID != n.actor {
				users = append(users, member.UserID)
			}
		}
	}
	if len(users) == 0 {
		return
	}
	devices, err := s.store.forUsers(ctx, users)
	if err != nil {
		log.Print("push registration lookup failed")
		return
	}
	for _, device := range devices {
		if ctx.Err() != nil {
			return
		}
		s.send(ctx, n, device)
	}
}

func (s *Sender) send(ctx context.Context, n notification, device models.PushDevice) {
	data := map[string]string{
		"kind": n.kind, "accountId": device.UserID.String(), "targetId": n.targetID.String(),
	}
	payload, _ := json.Marshal(map[string]any{
		"message": map[string]any{
			"token": device.Token, "data": data,
			"android": map[string]string{"priority": "high", "ttl": "600s"},
		},
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.url, bytes.NewReader(payload))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	response, err := s.client.Do(req)
	if err != nil {
		log.Print("FCM request failed")
		return
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusOK {
		return
	}
	var failure struct {
		Error struct {
			Details []struct {
				Code string `json:"errorCode"`
			} `json:"details"`
		} `json:"error"`
	}
	_ = json.NewDecoder(io.LimitReader(response.Body, 16*1024)).Decode(&failure)
	for _, detail := range failure.Error.Details {
		if detail.Code == "UNREGISTERED" {
			s.store.removeInvalid(ctx, device)
			return
		}
	}
	log.Printf("FCM rejected notification with HTTP %d", response.StatusCode)
}

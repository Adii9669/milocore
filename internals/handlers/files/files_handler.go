package files

import (
	"chat-server/internals/db/models"
	"chat-server/internals/middleware"
	"chat-server/internals/services"
	"chat-server/internals/transport/dto"
	"chat-server/internals/websockets"
	"chat-server/storage"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

const maxUploadSize int64 = 50 << 20

type FileHandler struct {
	storage  storage.FileStorage
	messages services.MessageService
	hub      *websockets.Hub
}

func NewFileHandler(fileStorage storage.FileStorage, messages services.MessageService, hub *websockets.Hub) *FileHandler {
	return &FileHandler{storage: fileStorage, messages: messages, hub: hub}
}

type uploadResponse struct {
	Message dto.MessageResponse `json:"message"`
}

func (h *FileHandler) Upload(w http.ResponseWriter, r *http.Request) {
	if h.storage == nil {
		http.Error(w, "file storage is not configured", http.StatusServiceUnavailable)
		return
	}
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize+(1<<20))
	file, header, err := r.FormFile("file")
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			http.Error(w, "file exceeds the 50 MB limit", http.StatusRequestEntityTooLarge)
		} else {
			http.Error(w, "file is required", http.StatusBadRequest)
		}
		return
	}
	defer file.Close()
	if header.Size > maxUploadSize {
		http.Error(w, "file exceeds the 50 MB limit", http.StatusRequestEntityTooLarge)
		return
	}

	filename := filepath.Base(strings.ReplaceAll(header.Filename, `\`, "/"))
	if filename == "" || filename == "." {
		http.Error(w, "invalid filename", http.StatusBadRequest)
		return
	}
	target, err := parseTarget(r, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	objectKey := "users/" + userID.String() + "/" + uuid.NewString()
	if err := h.storage.Upload(r.Context(), objectKey, file, header.Size, target.MimeType); err != nil {
		http.Error(w, "failed to upload file", http.StatusBadGateway)
		return
	}
	event, err := h.messages.CreateAttachmentMessage(r.Context(), userID, target.message, services.AttachmentInput{
		ObjectKey: objectKey,
		FileName:  filename,
		FileSize:  header.Size,
		MimeType:  target.MimeType,
	})
	if err != nil {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = h.storage.DeleteObject(cleanupCtx, objectKey)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if h.hub != nil {
		h.hub.PublishMessage(event)
	}

	isMine := true
	response := dto.MessageResponse{
		ID: event.ID.String(), Type: event.Type, Content: event.Content,
		CreatedAt: event.CreatedAt, IsMine: &isMine,
		Sender:  &dto.SenderDTO{ID: event.SenderID.String(), Name: event.SenderName},
		FileURL: event.FileURL, FilePreviewURL: event.FilePreviewURL, FileName: event.FileName, FileSize: event.FileSize, MimeType: event.MimeType,
	}
	if event.ChannelID != nil {
		channelID := event.ChannelID.String()
		response.ChannelID = &channelID
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(uploadResponse{Message: response})
}

type parsedTarget struct {
	message  services.IncomingMessage
	MimeType string
}

func parseTarget(r *http.Request, file multipart.File) (parsedTarget, error) {
	targetType := r.FormValue("type")
	caption := r.FormValue("caption")
	target := services.IncomingMessage{Type: targetType, Content: caption}
	switch targetType {
	case "dm":
		id, err := uuid.Parse(r.FormValue("receiverId"))
		if err != nil {
			return parsedTarget{}, errors.New("valid receiverId is required for a DM")
		}
		target.ReceiverID = &id
	case "channel":
		id, err := uuid.Parse(r.FormValue("channelId"))
		if err != nil {
			return parsedTarget{}, errors.New("valid channelId is required for a channel message")
		}
		target.ChannelID = &id
	default:
		return parsedTarget{}, errors.New("type must be dm or channel")
	}
	mimeType := "application/octet-stream"
	if seeker, ok := file.(io.Seeker); ok {
		buffer := make([]byte, 512)
		n, _ := file.Read(buffer)
		if n > 0 {
			mimeType = http.DetectContentType(buffer[:n])
		}
		_, _ = seeker.Seek(0, io.SeekStart)
	}
	return parsedTarget{message: target, MimeType: mimeType}, nil
}

func (h *FileHandler) Download(w http.ResponseWriter, r *http.Request) {
	if h.storage == nil {
		http.Error(w, "file storage is not configured", http.StatusServiceUnavailable)
		return
	}
	message, err := h.authorizedAttachment(w, r)
	if err != nil {
		return
	}
	url, err := h.storage.PresignedGetObject(r.Context(), *message.ObjectKey, dereference(message.FileName), "attachment", 5*time.Minute)
	if err != nil {
		http.Error(w, "failed to prepare download", http.StatusBadGateway)
		return
	}
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func dereference(value *string) string {
	if value == nil {
		return "attachment"
	}
	return *value
}

func (h *FileHandler) Preview(w http.ResponseWriter, r *http.Request) {
	message, err := h.authorizedAttachment(w, r)
	if err != nil {
		return
	}
	if message.MimeType == nil || !isPreviewableImage(*message.MimeType) {
		http.Error(w, "preview is not supported for this file type", http.StatusUnsupportedMediaType)
		return
	}
	url, err := h.storage.PresignedGetObject(r.Context(), *message.ObjectKey, dereference(message.FileName), "inline", 5*time.Minute)
	if err != nil {
		http.Error(w, "failed to prepare preview", http.StatusBadGateway)
		return
	}
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *FileHandler) authorizedAttachment(w http.ResponseWriter, r *http.Request) (*models.Message, error) {
	if h.storage == nil {
		http.Error(w, "file storage is not configured", http.StatusServiceUnavailable)
		return nil, errors.New("file storage unavailable")
	}
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return nil, err
	}
	messageID, err := uuid.Parse(chi.URLParam(r, "messageID"))
	if err != nil {
		http.Error(w, "invalid message ID", http.StatusBadRequest)
		return nil, err
	}
	message, err := h.messages.GetAttachment(r.Context(), userID, messageID)
	if err != nil || message == nil || message.ObjectKey == nil {
		http.NotFound(w, r)
		return nil, errors.New("attachment not found")
	}
	return message, nil
}

func isPreviewableImage(mimeType string) bool {
	switch strings.ToLower(strings.TrimSpace(strings.SplitN(mimeType, ";", 2)[0])) {
	case "image/jpeg", "image/png", "image/gif", "image/webp", "image/avif", "image/bmp":
		return true
	default:
		return false
	}
}

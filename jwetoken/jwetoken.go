package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-jose/go-jose/v4"
)

// Claims defines the structure of the data inside the token.
type JWEClaims struct {
	Sub  string `json:"sub"`
	Name string `json:"name"`
	Exp  int64  `json:"exp"`
}

type TokenPayLoadProvider interface {
	GetID() string
	GetName() *string
}

func GenerateEncryptedToken(provider TokenPayLoadProvider) (string, error) {
	//take the token key
	keyBase64 := os.Getenv("TOKEN_KEY")
	if keyBase64 == "" {
		return "", fmt.Errorf("TOKEN_KEY Not Set")
	}

	//encode it in base64
	key, err := base64.StdEncoding.DecodeString(keyBase64)
	if err != nil {
		return "", fmt.Errorf("invlaid key :%w ", err)
	}

	//create a encrypter for encrypting the  key (custom)
	encrypter, err := jose.NewEncrypter(
		jose.A256GCM,
		jose.Recipient{Algorithm: jose.DIRECT, Key: key},
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("Failed to Create Encrypter: %w", err)
	}

	//create the claims for that user to send the data
	claims := JWEClaims{
		Sub:  provider.GetID(),
		Name: *provider.GetName(),
		Exp:  time.Now().Add(1 * time.Hour).Unix(),
	}

	//take the claims in the json marshal and make it redable in json format
	plaintext, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("Failed To Marshal Claims: %w", err)
	}

	// take that plaintext and encrypt it to hide the deatails from an unknow users
	jwe, err := encrypter.Encrypt(plaintext)
	if err != nil {
		return "", fmt.Errorf("Failed to Encrypt: %w", err)
	}

	return jwe.CompactSerialize()

}

// ValidateToken parses and decrypts a JWE token string.
// It returns the user claims if the token is valid, otherwise it returns an error.
func ValidateEncryptedToken(tokenString string) (*JWEClaims, error) {

	//get the key
	keyBase64 := os.Getenv("TOKEN_KEY")
	//Decode the key
	key, _ := base64.StdEncoding.DecodeString(keyBase64)

	log.Println("Validation Begins")
	//validate the decoded key
	jwe, err := jose.ParseEncrypted(
		tokenString,
		[]jose.KeyAlgorithm{jose.DIRECT},
		[]jose.ContentEncryption{jose.A256GCM},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %v", err)
	}

	// 3. Decrypt the token using the key.
	plaintext, err := jwe.Decrypt(key)
	if err != nil {
		log.Printf("VALIDATOR: Decrypt Failed  %v", plaintext)
		return nil, fmt.Errorf("failed to decrypt token: %v", err)
	}

	// 5. Unmarshal the decrypted JSON payload into the Claims struct.
	var claims JWEClaims
	if err := json.Unmarshal(plaintext, &claims); err != nil {
		log.Printf("Middleware Error: %v", err)
		return nil, fmt.Errorf("failed to unmarshal claims: %v", err)
	}

	if time.Now().Unix() > claims.Exp {
		return nil, fmt.Errorf("Token is Expired")
	}

	log.Println("------Claims Done--------------")
	return &claims, nil
}

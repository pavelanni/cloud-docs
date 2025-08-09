package token

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Token struct {
	ID        string    `json:"id"`
	ExpiresAt time.Time `json:"expires_at"`
	IssuedAt  time.Time `json:"issued_at"`
}

type Manager struct {
	secret []byte
}

func NewManager(secret string) *Manager {
	return &Manager{
		secret: []byte(secret),
	}
}

func (m *Manager) Generate(ttl time.Duration) (string, error) {
	now := time.Now().UTC()
	token := Token{
		ID:        uuid.New().String(),
		ExpiresAt: now.Add(ttl),
		IssuedAt:  now,
	}

	payload, err := json.Marshal(token)
	if err != nil {
		return "", fmt.Errorf("failed to marshal token: %w", err)
	}

	encodedPayload := base64.URLEncoding.EncodeToString(payload)
	signature := m.sign(encodedPayload)
	encodedSignature := base64.URLEncoding.EncodeToString(signature)

	return fmt.Sprintf("%s.%s", encodedPayload, encodedSignature), nil
}

func (m *Manager) Validate(tokenString string) (*Token, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid token format")
	}

	encodedPayload := parts[0]
	encodedSignature := parts[1]

	expectedSignature := m.sign(encodedPayload)
	providedSignature, err := base64.URLEncoding.DecodeString(encodedSignature)
	if err != nil {
		return nil, fmt.Errorf("invalid signature encoding: %w", err)
	}

	if !hmac.Equal(expectedSignature, providedSignature) {
		return nil, fmt.Errorf("invalid token signature")
	}

	payload, err := base64.URLEncoding.DecodeString(encodedPayload)
	if err != nil {
		return nil, fmt.Errorf("invalid payload encoding: %w", err)
	}

	var token Token
	if err := json.Unmarshal(payload, &token); err != nil {
		return nil, fmt.Errorf("invalid token payload: %w", err)
	}

	if time.Now().UTC().After(token.ExpiresAt) {
		return nil, fmt.Errorf("token has expired")
	}

	return &token, nil
}

func (m *Manager) sign(data string) []byte {
	h := hmac.New(sha256.New, m.secret)
	h.Write([]byte(data))
	return h.Sum(nil)
}

func ParseDuration(s string) (time.Duration, error) {
	if s == "" {
		return 24 * time.Hour, nil
	}

	if duration, err := time.ParseDuration(s); err == nil {
		return duration, nil
	}

	if hours, err := strconv.Atoi(s); err == nil {
		return time.Duration(hours) * time.Hour, nil
	}

	return 0, fmt.Errorf("invalid duration format: %s", s)
}
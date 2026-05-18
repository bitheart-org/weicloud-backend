package service

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type vncToken struct {
	UserID    string
	VMID      string
	ExpiresAt time.Time
}

type VNCService struct {
	mu        sync.Mutex
	tokens    map[string]vncToken
	vmSession map[string]string
}

func NewVNCService() *VNCService {
	return &VNCService{
		tokens:    make(map[string]vncToken),
		vmSession: make(map[string]string),
	}
}

func (s *VNCService) IssueToken(userID, vmID string, ttl time.Duration) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	token := uuid.NewString()
	s.tokens[token] = vncToken{
		UserID:    userID,
		VMID:      vmID,
		ExpiresAt: time.Now().Add(ttl),
	}
	return token
}

func (s *VNCService) ConsumeToken(token, vmID string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.tokens[token]
	if !ok {
		return "", fmt.Errorf("invalid token")
	}
	delete(s.tokens, token)

	if entry.VMID != vmID {
		return "", fmt.Errorf("token vm mismatch")
	}
	if time.Now().After(entry.ExpiresAt) {
		return "", fmt.Errorf("token expired")
	}
	if activeToken := s.vmSession[vmID]; activeToken != "" {
		return "", fmt.Errorf("vm already has active vnc session")
	}
	s.vmSession[vmID] = token
	return entry.UserID, nil
}

func (s *VNCService) ReleaseSession(vmID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.vmSession, vmID)
}

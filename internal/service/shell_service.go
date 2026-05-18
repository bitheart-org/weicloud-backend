package service

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type shellToken struct {
	UserID    string
	VMID      string
	ExpiresAt time.Time
}

type ShellService struct {
	mu        sync.Mutex
	tokens    map[string]shellToken
	vmSession map[string]string
}

func NewShellService() *ShellService {
	return &ShellService{
		tokens:    make(map[string]shellToken),
		vmSession: make(map[string]string),
	}
}

func (s *ShellService) IssueToken(userID, vmID string, ttl time.Duration) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	token := uuid.NewString()
	s.tokens[token] = shellToken{
		UserID:    userID,
		VMID:      vmID,
		ExpiresAt: time.Now().Add(ttl),
	}
	return token
}

func (s *ShellService) ConsumeToken(token, vmID string) (string, error) {
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
		return "", fmt.Errorf("vm already has active shell session")
	}
	s.vmSession[vmID] = token
	return entry.UserID, nil
}

func (s *ShellService) ReleaseSession(vmID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.vmSession, vmID)
}

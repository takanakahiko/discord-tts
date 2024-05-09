package session

import (
	"errors"
	"fmt"
)

var (
	errTtsSessionManager  = errors.New("TtsSessionManager error")
	ErrTtsSessionNotFound = errors.New("ttsSession not found")
)

type TtsSessionManager struct {
	sessions []*TtsSession
}

// NewTtsSession create new TtsSessionManager.
func NewTtsSessionManager() *TtsSessionManager {
	return &TtsSessionManager{
		sessions: []*TtsSession{},
	}
}

// GetByGuildID.
func (t *TtsSessionManager) GetByGuildID(guildID string) (*TtsSession, error) {
	for _, s := range t.sessions {
		if s.GuildID() == guildID {
			return s, nil
		}
	}
	return nil, ErrTtsSessionNotFound
}

// Add.
func (t *TtsSessionManager) Add(ttsSession *TtsSession) error {
	_, err := t.GetByGuildID(ttsSession.GuildID())
	if !errors.Is(err, ErrTtsSessionNotFound) {
		return fmt.Errorf("ttsSession is already in voice-chat: %w", errTtsSessionManager)
	}
	t.sessions = append(t.sessions, ttsSession)
	return nil
}

// Remove.
func (t *TtsSessionManager) Remove(guildID string) error {
	ret := make([]*TtsSession, 0, len(t.sessions)-1)
	for _, v := range t.sessions {
		if v.GuildID() != guildID {
			ret = append(ret, v)
		}
	}
	t.sessions = ret
	return nil
}

package session

import (
	"fmt"
)

var (
	ErrTtsSessionNotFound = fmt.Errorf("ttsSession not found")
)

type TtsSessionManager struct {
	sessions []*TtsSession
}

// NewTtsSession create new TtsSessionManager
func NewTtsSessionManager() *TtsSessionManager {
	return &TtsSessionManager{
		sessions: []*TtsSession{},
	}
}

// GetByGuidID
func (t *TtsSessionManager) GetByGuidID(guidID string) (*TtsSession, error) {
	for _, s := range t.sessions {
		if s.GuidID() == guidID {
			return s, nil
		}
	}
	return nil, ErrTtsSessionNotFound
}

// Add
func (t *TtsSessionManager) Add(ttsSession *TtsSession) error {
	_, err := t.GetByGuidID(ttsSession.GuidID())
	if err != ErrTtsSessionNotFound {
		return fmt.Errorf("ttsSession is already in voice-chat")
	}
	t.sessions = append(t.sessions, ttsSession)
	return nil
}

// Remove
func (t *TtsSessionManager) Remove(guidID string) error {
	var ret []*TtsSession
	for _, v := range t.sessions {
		if v.GuidID() == guidID {
			continue
		}
		ret = append(ret, v)
	}
	t.sessions = ret
	return nil
}

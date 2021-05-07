package session

import (
	"fmt"
)

var(
	ErrTtsSessionNotFound = fmt.Errorf("ttsSession not found")
)

type TtsSessionManager struct {
	sessions []*TtsSession
}

func NewTtsSessionManager() *TtsSessionManager {
	return &TtsSessionManager{
		sessions: []*TtsSession{},
	}
}

// GetByTextChannelID
func (t *TtsSessionManager) GetByTextChannelID(textChannelID string) (*TtsSession, error) {
	for _, v := range t.sessions {
		if v.TextChanelID == textChannelID {
			return v, nil
		}
	}
	return nil, ErrTtsSessionNotFound
}

// GetByVoiceChannelID
func (t *TtsSessionManager) GetByVoiceChannelID(voiceChannelID string) (*TtsSession, error) {
	for _, v := range t.sessions {
		if v.VoiceConnection.ChannelID == voiceChannelID {
			return v, nil
		}
	}
	return nil, ErrTtsSessionNotFound
}

// Add
func (t *TtsSessionManager) Add(ttsSession *TtsSession) error {
	_, err := t.GetByTextChannelID(ttsSession.TextChanelID)
	if err != ErrTtsSessionNotFound {
		return fmt.Errorf("ttsSession is already in voice-chat")
	}
	t.sessions = append(t.sessions, ttsSession)
	return nil
}

// Remove
func (t *TtsSessionManager) Remove(textChannelID string) error {
	var ret []*TtsSession
	for _, v := range t.sessions {
		if v.TextChanelID == textChannelID {
			continue
		}
		ret = append(ret, v)
	}
	t.sessions = ret
	return nil
}

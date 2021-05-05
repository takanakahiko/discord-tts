package session

import (
	"sync"

	"github.com/bwmarrin/discordgo"
)

// TtsSession is a data structure for managing bot agents that participate in one voice channel
type TtsSession struct {
	TextChanelID    string
	VoiceConnection *discordgo.VoiceConnection
	Mut             sync.Mutex
	SpeechSpeed     float32
}

// NewTtsSession create new TtsSession
func NewTtsSession() *TtsSession {
	return &TtsSession{
		TextChanelID:    "",
		VoiceConnection: nil,
		Mut:             sync.Mutex{},
		SpeechSpeed:     1.0,
	}
}

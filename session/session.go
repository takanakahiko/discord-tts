package session

import (
	"fmt"
	"log"
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

// SendMessage send text to text chat
func (t *TtsSession) SendMessage(discord *discordgo.Session, format string, v ...interface{}) {
	if t.TextChanelID == "" {
		log.Println("Error sending message: TextChanelID is not set")
	}
	msg := fmt.Sprintf(format, v...)
	_, err := discord.ChannelMessageSend(t.TextChanelID, "[BOT] "+msg)
	log.Println(">>> " + msg)
	if err != nil {
		log.Println("Error sending message: ", err)
	}
}

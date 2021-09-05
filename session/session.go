package session

import (
	"fmt"
	"io"
	"log"
	"net/url"
	"regexp"
	"strconv"
	"sync"
	"time"

	"golang.org/x/text/language"

	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
)

// TtsSession is a data structure for managing bot agents that participate in one voice channel
type TtsSession struct {
	TextChanelID    string
	VoiceConnection *discordgo.VoiceConnection
	mut             sync.Mutex
	speechSpeed     float64
	speechLanguage  string
	guildID         string
}

// NewTtsSession create new TtsSession
func NewTtsSession() *TtsSession {
	return &TtsSession{
		TextChanelID:    "",
		VoiceConnection: nil,
		mut:             sync.Mutex{},
		speechSpeed:     1.5,
		speechLanguage:  "auto",
	}
}

// GetByGuidID
func (t *TtsSession) GuidID() string {
	return t.guildID
}

// Get state of VoiceConnection
func (t *TtsSession) IsConnected() bool {
	return t.VoiceConnection != nil && t.VoiceConnection.Ready
}

// Join join the same channel as the caller
func (t *TtsSession) Join(discord *discordgo.Session, callerUserID, textChannelID string) error {
	if t.VoiceConnection != nil {
		return fmt.Errorf("bot is already in voice-chat")
	}

	var callUserVoiceState *discordgo.VoiceState = nil
	for _, guild := range discord.State.Guilds {
		for _, vs := range guild.VoiceStates {
			if vs.UserID == callerUserID {
				callUserVoiceState = vs
			}
		}
	}
	if callUserVoiceState == nil {
		t.SendMessage(discord, "Caller is not in voice-chat.")
		return fmt.Errorf("caller is not in voice-chat")
	}

	voiceConnection, err := discord.ChannelVoiceJoin(
		callUserVoiceState.GuildID, callUserVoiceState.ChannelID, false, true)
	if err != nil {
		t.SendMessage(discord, err.Error())
		return fmt.Errorf(
			"failed ChannelVoiceJoin(gID=%s, cID=%s, mute=false, deaf=true): %w",
			callUserVoiceState.GuildID, callUserVoiceState.ChannelID, err)
	}
	t.VoiceConnection = voiceConnection
	t.TextChanelID = textChannelID
	t.guildID = voiceConnection.GuildID
	t.SendMessage(discord, "Joined to voice chat!\n speechSpeed:%g\n speechLanguage:%s", t.speechSpeed, t.speechLanguage)
	return nil
}

// SendMessage send text to text chat
func (t *TtsSession) SendMessage(discord *discordgo.Session, format string, v ...interface{}) {
	if t.TextChanelID == "" {
		log.Println("Error sending message: TextChanelID is not set")
	}
	msg := fmt.Sprintf(format, v...)
	log.Println(">>> " + msg)
	if _, err := discord.ChannelMessageSend(t.TextChanelID, "[BOT] "+msg); err != nil {
		log.Println("Error sending message: ", err)
	}
}

// Speech speech the received text on the voice channel
func (t *TtsSession) Speech(discord *discordgo.Session, text string) error {
	if regexp.MustCompile(`<a:|<@|<#|<@&|http|` + "```").MatchString(text) {
		t.SendMessage(discord, "Skipped reading")
		return fmt.Errorf("text is emoji, mention channel, group mention or url")
	}

	text = regexp.MustCompile(`<:(.+?):[0-9]+>`).ReplaceAllString(text, "$1")
	text = regexp.MustCompile(`_`).ReplaceAllString(text, "")

	lang := t.speechLanguage
	if lang == "auto" {
		lang = "ja"
		if regexp.MustCompile(`^[a-zA-Z0-9\s.,]+$`).MatchString(text) {
			lang = "en"
		}
	}

	t.mut.Lock()
	defer t.mut.Unlock()

	voiceURL := fmt.Sprintf("http://translate.google.com/translate_tts?ie=UTF-8&textlen=32&client=tw-ob&q=%s&tl=%s", url.QueryEscape(text), lang)
	if err := t.playAudioFile(voiceURL); err != nil {
		t.SendMessage(discord, "err=%s", err.Error())
		return fmt.Errorf("t.playAudioFile(voiceURL:%+v) fail: %w", voiceURL, err)
	}
	return nil
}

// Leave end connection and init variables
func (t *TtsSession) Leave(discord *discordgo.Session) error {
	if err := t.VoiceConnection.Disconnect(); err != nil {
		return fmt.Errorf("t.VoiceConnection.Disconnect() fail: %w", err)
	}
	t.SendMessage(discord, "Left from voice chat...")
	t.VoiceConnection = nil
	t.TextChanelID = ""
	return nil
}

// SetSpeechSpeed validate and set speechSpeed
func (t *TtsSession) SetSpeechSpeed(discord *discordgo.Session, newSpeechSpeed float64) error {
	if newSpeechSpeed < 0.5 || newSpeechSpeed > 100 {
		t.SendMessage(discord, "You can set a value from 0.5 to 100")
		return fmt.Errorf("newSpeechSpeed=%v is invalid", newSpeechSpeed)
	}
	t.speechSpeed = newSpeechSpeed
	t.SendMessage(discord, "Changed speed to %s", strconv.FormatFloat(newSpeechSpeed, 'f', -1, 64))
	return nil
}

// SetLanguage
func (t *TtsSession) SetLanguage(discord *discordgo.Session, langText string) error {
	if langText == "auto" {
		t.speechLanguage = langText
		t.SendMessage(discord, "Changed language to '%s'", t.speechLanguage)
		return nil
	}

	if lang, err := language.Parse(langText); err != nil {
		return fmt.Errorf("Language.Parse() fail: %w", err)
	} else {
		t.speechLanguage = lang.String()
	}

	t.SendMessage(discord, "Changed language to '%s'", t.speechLanguage)
	return nil
}

// playAudioFile play audio file on the voice channel
func (t *TtsSession) playAudioFile(filename string) error {
	if err := t.VoiceConnection.Speaking(true); err != nil {
		return fmt.Errorf("t.VoiceConnection.Speaking(true) fail: %w", err)
	}
	defer func() {
		if err := t.VoiceConnection.Speaking(false); err != nil {
			log.Fatal(err)
		}
	}()

	opts := dca.StdEncodeOptions
	opts.CompressionLevel = 0
	opts.RawOutput = true
	opts.Bitrate = 120
	opts.AudioFilter = fmt.Sprintf("atempo=%f", t.speechSpeed)

	encodeSession, err := dca.EncodeFile(filename, opts)
	if err != nil {
		return fmt.Errorf("dca.EncodeFile(filename:%+v, opts:%+v) fail: %w", filename, opts, err)
	}

	done := make(chan error)
	stream := dca.NewStream(encodeSession, t.VoiceConnection, done)
	ticker := time.NewTicker(time.Second)

	for {
		select {
		case err := <-done:
			if err != nil && err != io.EOF {
				return err
			}
			encodeSession.Cleanup()
			return nil
		case <-ticker.C:
			stats := encodeSession.Stats()
			playbackPosition := stream.PlaybackPosition()
			log.Printf("Sending Now... : Playback: %10s, Transcode Stats: Time: %5s, Size: %5dkB, Bitrate: %6.2fkB, Speed: %5.1fx\r", playbackPosition, stats.Duration.String(), stats.Size, stats.Bitrate, stats.Speed)
		}
	}
}

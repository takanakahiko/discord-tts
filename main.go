package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
)

var (
	textChanelID string                     = "not set"
	vcsession    *discordgo.VoiceConnection = nil
	mut          sync.Mutex
)

func main() {
	discord, err := discordgo.New()
	discord.Token = "Bot " + os.Getenv("TOKEN")
	if err != nil {
		fmt.Println("Error logging in")
		fmt.Println(err)
	}

	discord.AddHandler(onMessageCreate)

	err = discord.Open()
	if err != nil {
		fmt.Println(err)
	}
	defer discord.Close()

	fmt.Println("Listening...")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	return
}

func clientID() string {
	return os.Getenv("CLIENT_ID")
}

func botName() string {
	return "<@!" + clientID() + ">"
}

func onMessageCreate(discord *discordgo.Session, m *discordgo.MessageCreate) {
	discordChannel, err := discord.Channel(m.ChannelID)
	if err != nil {
		log.Printf("\t%s\t%s\t>\t%s\n", m.ChannelID, m.Author.Username, m.Content)
	} else {
		log.Printf("\t%s\t%s\t>\t%s\n", discordChannel.Name, m.Author.Username, m.Content)
	}

	switch {
	case strings.HasPrefix(m.Content, botName()+" join"):
		var err error
		vcsession, err = joinUserVoiceChannel(discord, m.Author.ID)
		if err != nil {
			sendMessage(discord, m.ChannelID, err.Error())
		}
		textChanelID = m.ChannelID
		sendMessage(discord, m.ChannelID, "Joined to voice chat!")
	case strings.HasPrefix(m.Content, botName()+" leave"):
		if vcsession == nil {
			return
		}
		err := vcsession.Disconnect()
		if err != nil {
			sendMessage(discord, m.ChannelID, err.Error())
		}
		sendMessage(discord, m.ChannelID, "Left from voice chat...")
		vcsession = nil
	case strings.Contains(m.Content, "http"):
		if vcsession != nil {
			sendMessage(discord, m.ChannelID, "URLなのでスキップしました")
		}
	case vcsession != nil && m.ChannelID == textChanelID && m.Author.ID != clientID():
		if vcsession != nil {
			mut.Lock()
			defer mut.Unlock()

			url := fmt.Sprintf("http://translate.google.com/translate_tts?ie=UTF-8&total=1&idx=0&textlen=32&client=tw-ob&q=%s&tl=%s", url.QueryEscape(m.Content), "ja")
			if err := playAudioFile(vcsession, url); err != nil {
				sendMessage(discord, m.ChannelID, err.Error())
			}
		}
	}
}

func sendMessage(discord *discordgo.Session, channelID string, msg string) {
	_, err := discord.ChannelMessageSend(channelID, "[BOT] "+msg)

	log.Println(">>> " + msg)
	if err != nil {
		log.Println("Error sending message: ", err)
	}
}

func joinUserVoiceChannel(discord *discordgo.Session, userID string) (*discordgo.VoiceConnection, error) {
	vs, err := findUserVoiceState(discord, userID)
	if err != nil {
		return nil, err
	}
	return discord.ChannelVoiceJoin(vs.GuildID, vs.ChannelID, false, true)
}

func findUserVoiceState(discord *discordgo.Session, userid string) (*discordgo.VoiceState, error) {
	for _, guild := range discord.State.Guilds {
		for _, vs := range guild.VoiceStates {
			if vs.UserID == userid {
				return vs, nil
			}
		}
	}
	return nil, errors.New("Could not find user's voice state")
}

func playAudioFile(v *discordgo.VoiceConnection, filename string) error {
	if err := v.Speaking(true); err != nil {
		return err
	}
	defer v.Speaking(false)

	opts := dca.StdEncodeOptions
	opts.RawOutput = true
	opts.Bitrate = 120

	encodeSession, err := dca.EncodeFile(filename, opts)
	if err != nil {
		return err
	}

	done := make(chan error)
	stream := dca.NewStream(encodeSession, v, done)
	ticker := time.NewTicker(time.Second)

	for {
		select {
		case err := <-done:
			if err != nil && err != io.EOF {
				return err
			}
			encodeSession.Truncate()
			return nil
		case <-ticker.C:
			stats := encodeSession.Stats()
			playbackPosition := stream.PlaybackPosition()
			log.Printf("Sending Now... : Playback: %10s, Transcode Stats: Time: %5s, Size: %5dkB, Bitrate: %6.2fkB, Speed: %5.1fx\r", playbackPosition, stats.Duration.String(), stats.Size, stats.Bitrate, stats.Speed)
		}
	}
}

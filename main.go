package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"regexp"

	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"

	"flag"
)

var (
	textChanelID = "not set"
	vcsession *discordgo.VoiceConnection = nil
	mut sync.Mutex
	speechSpeed float32 = 1.0
	prefix = flag.String("prefix", "", "call prefix")
)

func main() {
	flag.Parse()
	fmt.Println("prefix       :",*prefix)
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
	// if prefix is "", you can call by mention
	if *prefix == "mention" {
		return "<@" + clientID() + ">"
	}
	return *prefix
}

func onMessageCreate(discord *discordgo.Session, m *discordgo.MessageCreate) {
	discordChannel, err := discord.Channel(m.ChannelID)
	if err != nil {
		log.Fatal(err)
		return
	} else {
		log.Printf("ch:%s user:%s > %s\n", discordChannel.Name, m.Author.Username, m.Content)
	}

	// bot check
	if m.Author.Bot {
		return
	}

	// "join" command
	if isCommandMessage(m.Content, "join") {
		if vcsession != nil {
			sendMessage(discord, m.ChannelID, "Bot is already in voice-chat.")
			return
		}
		vcsession, err = joinUserVoiceChannel(discord, m.Author.ID)
		if err != nil {
			sendMessage(discord, m.ChannelID, err.Error())
			return
		}
		textChanelID = m.ChannelID
		sendMessage(discord, m.ChannelID, "Joined to voice chat!")
		return
	}

	// ignore case of "not join", "another channel" or "include ignore prefix"
	if vcsession == nil || m.ChannelID != textChanelID || strings.HasPrefix(m.Content, ";") {
		return
	}

	// other commands
	switch {
	case isCommandMessage(m.Content, "leave"):
		err := vcsession.Disconnect()
		if err != nil {
			sendMessage(discord, m.ChannelID, err.Error())
		}
		sendMessage(discord, m.ChannelID, "Left from voice chat...")
		vcsession = nil
	case isCommandMessage(m.Content, "speed"):
		speedStr := strings.Replace(m.Content, botName()+" speed ", "", 1)
		if newSpeed, err := strconv.ParseFloat(speedStr, 32); err == nil {
			speechSpeed = float32(newSpeed)
			sendMessage(discord, m.ChannelID, fmt.Sprintf("速度を%sに変更しました", strconv.FormatFloat(newSpeed, 'f', -1, 32)))
		}
	}

	// ignore emoji, mention channel, group mention and url
	if regexp.MustCompile(`<a:|<@|<#|<@&|http`).MatchString(m.Content) {
		sendMessage(discord, m.ChannelID, "読み上げをスキップしました")
	}

	// Speech
	mut.Lock()
	defer mut.Unlock()
	url := fmt.Sprintf("http://translate.google.com/translate_tts?ie=UTF-8&textlen=32&client=tw-ob&q=%s&tl=%s", url.QueryEscape(m.Content), "ja")
	if err := playAudioFile(vcsession, url); err != nil {
		sendMessage(discord, m.ChannelID, err.Error())
	}
}

func isCommandMessage(message, command string) bool {
	return strings.HasPrefix(message, botName()+" "+command)
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
	opts.AudioFilter = fmt.Sprintf("atempo=%f", speechSpeed)

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

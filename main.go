package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/takanakahiko/discord-tts/logger"
	"github.com/takanakahiko/discord-tts/session"
)

var (
	sessionManager = session.NewTtsSessionManager()
	prefix         = flag.String("prefix", "", "call prefix")
	clientID       = ""
)

func main() {
	flag.Parse()
	fmt.Println("prefix       :", *prefix)

	discord, err := discordgo.New()
	if err != nil {
		fmt.Println("Error logging in")
		fmt.Println(err)
	}

	discord.Token = "Bot " + os.Getenv("TOKEN")
	discord.AddHandler(onReady)
	discord.AddHandler(onMessageCreate)
	discord.AddHandler(onVoiceStateUpdate)

	if err = discord.Open(); err != nil {
		fmt.Println(err)
	}
	defer func() {
		if err := discord.Close(); err != nil {
			logger.PrintError(err)
		}
	}()

	fmt.Println("Listening...")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}

func botName() string {
	// if prefix is "", you can call by mention
	if *prefix == "mention" {
		return "<@" + clientID + ">"
	}
	return *prefix
}

func onReady(discord *discordgo.Session, r *discordgo.Ready) {
	clientID = discord.State.User.ID
}

//event by message
func onMessageCreate(discord *discordgo.Session, m *discordgo.MessageCreate) {

	{
		discordChannel, err := discord.Channel(m.ChannelID)
		if err != nil {
			log.Fatal(err)
			return
		}
		guild, err := discord.Guild(m.GuildID)
		if err != nil && err != session.ErrTtsSessionNotFound {
			log.Println(err)
			return
		}
		log.Printf("onMessageCreate\n server: %s\n ch: %s\n user: %s\n message: %s\n", guild.Name, discordChannel.Name, m.Author.Username, m.Content)
	}

	// bot check
	if m.Author.Bot || strings.HasPrefix(m.Content, ";") {
		return
	}

	// "join" command
	if isCommandMessage(m.Content, "join") {
		if _, err := sessionManager.GetByGuidID(m.GuildID); err == nil {
			sendMessage(discord, m.ChannelID, "Bot is already in voice-chat.")
			return
		} else if err != session.ErrTtsSessionNotFound {
			log.Println(err)
			return
		}
		ttsSession := session.NewTtsSession()
		if err := ttsSession.Join(discord, m.Author.ID, m.ChannelID); err != nil {
			logger.PrintError(err)
			return
		}
		if err := sessionManager.Add(ttsSession); err != nil {
			logger.PrintError(err)
		}
		return
	}

	// ignore case of "not join" or "include ignore prefix"
	ttsSession, err := sessionManager.GetByGuidID(m.GuildID)
	if err == session.ErrTtsSessionNotFound {
		return
	}
	if err != nil {
		logger.PrintError(err)
		return
	}

	// Ignore if the TextChanelID of session and the channel of the message are different
	if ttsSession.TextChanelID != m.ChannelID {
		return
	}

	// other commands
	switch {
	case isCommandMessage(m.Content, "leave"):
		if err := ttsSession.Leave(discord); err != nil {
			logger.PrintError(err)
		}
		if err := sessionManager.Remove(ttsSession.GuidID()); err != nil {
			logger.PrintError(err)
		}
		return
	case isCommandMessage(m.Content, "speed"):
		speedStr := strings.Replace(m.Content, botName()+" speed ", "", 1)
		newSpeed, err := strconv.ParseFloat(speedStr, 64)
		if err != nil {
			ttsSession.SendMessage(discord, "数字ではない値は設定できません")
			return
		}
		if err = ttsSession.SetSpeechSpeed(discord, newSpeed); err != nil {
			logger.PrintError(err)
		}
		return
	case isCommandMessage(m.Content, "lang"):
		newLang := strings.Replace(m.Content, botName()+" lang ", "", 1)
		if err = ttsSession.SetLanguage(discord, newLang); err != nil {
			logger.PrintError(err)
		}
		return
	}

	if err = ttsSession.Speech(discord, m.Content); err != nil {
		log.Println(err)
	}
}

func onVoiceStateUpdate(discord *discordgo.Session, v *discordgo.VoiceStateUpdate) {
	ttsSession, err := sessionManager.GetByGuidID(v.GuildID)
	if err == session.ErrTtsSessionNotFound {
		return
	}
	if err != nil {
		log.Println(err)
		return
	}

	if !ttsSession.IsConnected() {
		return
	}

	// ボイスチャンネルに誰かしらいたら return
	for _, guild := range discord.State.Guilds {
		for _, vs := range guild.VoiceStates {
			if ttsSession.VoiceConnection.ChannelID == vs.ChannelID && vs.UserID != clientID {
				return
			}
		}
	}

	// ボイスチャンネルに誰もいなかったら Disconnect する
	if err := sessionManager.Remove(v.GuildID); err != nil {
		log.Println(err)
	}
	if err = ttsSession.Leave(discord); err != nil {
		log.Println(err)
	}
}

func isCommandMessage(message, command string) bool {
	return strings.HasPrefix(message, botName()+" "+command)
}

func sendMessage(discord *discordgo.Session, textChanelID, format string, v ...interface{}) {
	session := session.NewTtsSession()
	session.TextChanelID = textChanelID
	session.SendMessage(discord, format, v...)
}

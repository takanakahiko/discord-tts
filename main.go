package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/takanakahiko/discord-tts/session"
)

var (
	ttsSession = session.NewTtsSession()
	prefix     = flag.String("prefix", "", "call prefix")
	clientID   = ""
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
	discord.AddHandler(onMessageCreate)
	discord.AddHandler(onVoiceStateUpdate)

	err = discord.Open()
	if err != nil {
		fmt.Println(err)
	}
	defer func() {
		if err := discord.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	fmt.Println("Listening...")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	return
}

func botName() string {
	// if prefix is "", you can call by mention
	if *prefix == "mention" {
		return "<@" + clientID + ">"
	}
	return *prefix
}

//event by message
func onMessageCreate(discord *discordgo.Session, m *discordgo.MessageCreate) {

	// main 内でやると、なぜかときどき失敗するので、確実に取得できそうな場所でやる
	// 確実に API が立たけるようになったタイミングでフックされる関数とかあったら、そこでやりたい
	clientID = discord.State.User.ID

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
		if ttsSession.VoiceConnection != nil {
			ttsSession.SendMessage(discord, "Bot is already in voice-chat.")
			return
		}
		ttsSession.VoiceConnection, err = joinUserVoiceChannel(discord, m.Author.ID)
		if err != nil {
			ttsSession.SendMessage(discord, err.Error())
			return
		}
		ttsSession.TextChanelID = m.ChannelID
		ttsSession.SendMessage(discord, "Joined to voice chat!")
		return
	}

	// ignore case of "not join", "another channel" or "include ignore prefix"
	if ttsSession.VoiceConnection == nil || m.ChannelID != ttsSession.TextChanelID || strings.HasPrefix(m.Content, ";") {
		return
	}

	// other commands
	switch {
	case isCommandMessage(m.Content, "leave"):
		err := ttsSession.VoiceConnection.Disconnect()
		if err != nil {
			ttsSession.SendMessage(discord, err.Error())
			return
		}
		ttsSession.SendMessage(discord, "Left from voice chat...")
		ttsSession.VoiceConnection = nil
		return
	case isCommandMessage(m.Content, "speed"):
		speedStr := strings.Replace(m.Content, botName()+" speed ", "", 1)

		speed, err := strconv.ParseFloat(speedStr, 64)
		if err != nil {
			ttsSession.SendMessage(discord, "数字ではない値は設定できません")
			return
		}
		if speed < 0.5 || speed > 100 {
			ttsSession.SendMessage(discord, "設定できるのは 0.5 ~ 100 の値です")
			return
		}

		if newSpeed, err := strconv.ParseFloat(speedStr, 32); err == nil {
			ttsSession.SpeechSpeed = float32(newSpeed)
			ttsSession.SendMessage(discord, "速度を%sに変更しました", strconv.FormatFloat(newSpeed, 'f', -1, 32))
		}
		return
	}

	if err = ttsSession.Speech(discord, m.Content); err != nil {
		log.Println(err)
	}
}

func onVoiceStateUpdate(discord *discordgo.Session, v *discordgo.VoiceStateUpdate) {
	if ttsSession.VoiceConnection == nil || !ttsSession.VoiceConnection.Ready {
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
	err := ttsSession.VoiceConnection.Disconnect()
	if err != nil {
		fmt.Printf("err=%+v", err)
	}
	ttsSession.VoiceConnection = nil

	if ttsSession.TextChanelID != "" {
		sendMessage(discord, ttsSession.TextChanelID, "Left from voice chat...")
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
	return nil, errors.New("could not find user's voice state")
}

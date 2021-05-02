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

	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"

	"flag"
)

var (
	textChanelID string                     = "not set"
	vcsession    *discordgo.VoiceConnection = nil
	mut          sync.Mutex
	speechSpeed  float32 = 1.0
	//call_type default:mention
	call_type = flag.String("call", "mention", "call prefix")
)

func main() {
	flag.Parse()
	fmt.Println("call         :",*call_type)
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
	//call_typeがメンション(default)ならばmentionで呼べるようreturn
	if (*call_type == "mention") {
		return "<@" + clientID() + ">"
	}
	//それ以外だったらcall_typeを先頭に設定してreturn
	return *call_type
}

//メッセージを受信して実行
func onMessageCreate(discord *discordgo.Session, m *discordgo.MessageCreate) {
	//メッセージの送られたチャンネルIDを定義
	discordChannel, err := discord.Channel(m.ChannelID)

	//メッセージをerrorがなければlogに転送?
	if err != nil {
		log.Printf("ch:%s user:%s > %s\n", m.ChannelID, m.Author.Username, m.Content)
	} else {
		log.Printf("ch:%s user:%s > %s\n", discordChannel.Name, m.Author.Username, m.Content)
	}

	//VC関連
	//bot check
	if m.Author.Bot {
		return
	}

	//join確認
	if strings.HasPrefix(m.Content, botName()+" join") {
		if vcsession != nil {
			sendMessage(discord, m.ChannelID, "Bot is already in voice-chat.")
			return
		}
		//エラー変数を定義
		var err error
		//VCに接続
		vcsession, err = joinUserVoiceChannel(discord, m.Author.ID)
		//正常に接続できなかったら実行
		if err != nil {
			sendMessage(discord, m.ChannelID, err.Error())
			return
		}
		//メッセージの送られたチャンネルIDを定義
		textChanelID = m.ChannelID
		//VCに接続したことを定義
		sendMessage(discord, m.ChannelID, "Joined to voice chat!")
		return
	}

	//VCセッション=nil OR 読み上げ!=送信チャンネル OR ;スタート
	if vcsession == nil || m.ChannelID != textChanelID || strings.HasPrefix(m.Content, ";") {
		return
	}

	//大半のメッセージ関連
	switch {
	//メッセージからleaveを検知して実行
	case strings.HasPrefix(m.Content, botName()+" leave"):
		//vcsessionをなくす
		err := vcsession.Disconnect()
		//errorがあれば表示
		if err != nil {
			sendMessage(discord, m.ChannelID, err.Error())
		}
		//VCから抜けたのをsendする
		sendMessage(discord, m.ChannelID, "Left from voice chat...")
		//VCにいないのを定義
		vcsession = nil
	//メッセージからspeedを検知して実行
	case strings.HasPrefix(m.Content, botName()+" speed "):
		//数字部分を切り出し
		speedStr := strings.Replace(m.Content, botName()+" speed ", "", 1)
		//切り出し結果が正しいか確認
		if newSpeed, err := strconv.ParseFloat(speedStr, 32); err == nil {
			//スピードを変数に設定
			speechSpeed = float32(newSpeed)
			//設定したことをstring to float して通知
			sendMessage(discord, m.ChannelID, fmt.Sprintf("速度を%sに変更しました", strconv.FormatFloat(newSpeed, 'f', -1, 32)))
		}
	//メッセージに<a: http <@ <# <@& が入ってないかを確認
	case regexp.MustCompile(`<a:|<@|<#|<@&|http`).MatchString(m.Content) :
		sendMessage(discord, m.ChannelID, "読み上げをスキップしました")
	//上に適合するものがない場合読み上げ
	default:
		//muteして二重発言を対策
		mut.Lock()
		//発言終わりまで待機
		defer mut.Unlock()
		//テキストをaudio file化
		url := fmt.Sprintf("http://translate.google.com/translate_tts?ie=UTF-8&textlen=32&client=tw-ob&q=%s&tl=%s", url.QueryEscape(m.Content), "ja")
		//playする
		if err := playAudioFile(vcsession, url); err != nil {
			//playに失敗したらerrorを表示
			sendMessage(discord, m.ChannelID, err.Error())
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

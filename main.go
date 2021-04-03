package main

import (
        "errors"
        "flag"
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
)

//botの初期設定(全ての部分において共通になる
var (
  textChanelID string                     = "not set"
  vcsession    *discordgo.VoiceConnection = nil
  mut          sync.Mutex
  speechSpeed  float32 = 1.0
  speechPitch  float32 = 1.0
  //botのcall token id の入手
  //call_type default:mention
  call_type = flag.String("call", "mention", "call prefix")
  //token default:unset
  token = flag.String("token", "", "bot token")
  //id default: unset
  id = flag.String("id", "", "bot client id")
)

//botのプログラム メイン部分(セットアップ discordからのメッセージ受信などの設定
func main() {
  //flagを入手&表示
  flag.Parse()
  fmt.Println("call         :",*call_type)
  fmt.Println("bot token    :",*token)
  fmt.Println("bot client id:",*id)
  //botのセッションを作成 起動
  discord, err := discordgo.New()
  discord.Token = "Bot " + *token
  //loginチェック
  if err != nil {
    fmt.Println("Error logging in")
    fmt.Println(err)
  }
  //メッセージの更新を受け取って呼び出すトリガーを作成
  discord.AddHandler(onMessageCreate)
  //discordに接続できるか確認
  err = discord.Open()
  if err != nil {
  fmt.Println(err)
  }
  //確認できるまで待機 確認出来たら処理を再開
  defer discord.Close()
  //discordのメッセージを読み込み始める
  fmt.Println("Listening...")

  //勝手に落ちないように起動しっぱなしにするシグナルを送信 停止トリガーは下を調べて
  sc := make(chan os.Signal, 1)
  signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
 <-sc

  return
}
//client id 設定
func clientID() string {
  //flagからclient idを取得して返り値として出す
  return *id
}

//先頭(旧mention)をどうするかをして指定
func botName() string {
  //返り値用の変数を定義
  var name string
  //call_typeがメンション(default)ならばmentionで呼べるように設定
  if (*call_type == "mention") {
    name = "<@" + clientID() + ">"
  } else {
    //それ以外だったらcall_typeをそのまま先頭に設定
    name = *call_type
  }
  //返り値を出す
  return name
}
//メッセージを受信して実行
func onMessageCreate(discord *discordgo.Session, m *discordgo.MessageCreate) {
  //メッセージの送られたチャンネルIDを定義
  discordChannel, err := discord.Channel(m.ChannelID)
  //メッセージをerrorがなければlogに転送?
  if err != nil {
    log.Printf("\t%s\t%s\t>\t%s\n", m.ChannelID, m.Author.Username, m.Content)
  } else {
    log.Printf("ch:%s user:%s > %s\n", discordChannel.Name, m.Author.Username, m.Content)
  }

  //VC接続していない
  if vcsession == nil {
    if strings.HasPrefix(m.Content, botName()+" join") {
      //エラー変数を定義
      var err error
      //VCに接続
      vcsession, err = joinUserVoiceChannel(discord, m.Author.ID)
      //正常に接続できなかったら実行
      if err != nil {
        sendMessage(discord, m.ChannelID, err.Error())
      }
      //メッセージの送られたチャンネルIDを定義
      textChanelID = m.ChannelID
      //VCに接続したことを定義
      sendMessage(discord, m.ChannelID, "Joined to voice chat!")
    }
  //VC接続済み textチャンネル=メッセージのチャンネル
  } else if m.ChannelID == textChanelID {
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
      //メッセージに ; が入ってないかを確認&switchで書かれてるからskip
      case strings.HasPrefix(m.Content, ";") || strings.HasPrefix(m.Content, "[BOT] "):
         log.Println("bot tts skip this message")
      //メッセージに<a: http <@ <# <@& が入ってないかを確認&switchで書かれてるからskip
      case strings.Contains(m.Content, "<a:") || strings.Contains(m.Content, "http") || strings.Contains(m.Content, "<@") || strings.Contains(m.Content, "<#") || strings.Contains(m.Content, "<@&"):
        sendMessage(discord, m.ChannelID, "読み上げをスキップしました")
      //text ch = メッセージ受信チャンネル,発言者 !=自分 かを確認
      case m.Author.ID != clientID():
        //muteして二重発言を対策
        mut.Lock()
        //発言終わりまで待機
        defer mut.Unlock()
        //テキストをaudio file化
        url := fmt.Sprintf("http://translate.google.com/translate_tts?ie=UTF-8&total=1&idx=0&textlen=32&client=tw-ob&q=%s&tl=%s", url.QueryEscape(m.Content), "ja")
        //playする
        if err := playAudioFile(vcsession, url); err != nil {
          //playに失敗したらerrorを表示
          sendMessage(discord, m.ChannelID, err.Error())
        }
    }
  }
}

//メッセージの送信
func sendMessage(discord *discordgo.Session, channelID string, msg string) {
  _, err := discord.ChannelMessageSend(channelID, "[BOT] "+msg)

  //コンソールにbotがメッセージを送ったことを表示
  log.Println("bot send > " + msg)
  //sendに失敗したときにerrorを表示
  if err != nil {
    log.Println("Error sending message: ", err)
  }
}

//ユーザーの入ってるVCに接続
func joinUserVoiceChannel(discord *discordgo.Session, userID string) (*discordgo.VoiceConnection, error) {
  //ユーザーの入ってるVCを検索
  vs, err := findUserVoiceState(discord, userID)
  if err != nil {
    return nil, err
  }
  return discord.ChannelVoiceJoin(vs.GuildID, vs.ChannelID, false, true)
}
//ユーザーの入ってるBCを検索
func findUserVoiceState(discord *discordgo.Session, userid string) (*discordgo.VoiceState, error) {
  //botの入っている全てのサーバーを検索
  for _, guild := range discord.State.Guilds {
    //そのサーバーのVCのステータスを確認
    for _, vs := range guild.VoiceStates {
      //ステータスのuserIDとメッセージを送ったuserIDが一致するかを確認
      if vs.UserID == userid {
        return vs, nil
      }
    }
  }
  //探しても見つからなかったときにerrorを表示
  return nil, errors.New("Could not find user's voice state")
}

//urlを通してできたaudiofileをplayする
func playAudioFile(v *discordgo.VoiceConnection, filename string) error {
  //喋ってる判定を発生
  if err := v.Speaking(true); err != nil {
    return err
  }
  //喋ってる判定がなくなるまで待機 そのあと 喋ってる判定を消す
  defer v.Speaking(false)

  //fileをエンコード
  opts := dca.StdEncodeOptions
  //speechSpeedを参照して再生速度を指定
  opts.AudioFilter = fmt.Sprintf("atempo=%f", speechSpeed)

  encodeSession, err := dca.EncodeFile(filename, opts)
  if err != nil {
    return err
  }

  done := make(chan error)
  stream := dca.NewStream(encodeSession, v, done)
  //タイマーを生成
  ticker := time.NewTicker(time.Second)
  //どっちにも属さなくなるまでloop
  for {
    select {
      case err := <-done:
        if err != nil && err != io.EOF {
          return err
        }
        encodeSession.Truncate()
        return nil
      case <-ticker.C:
        playbackPosition := stream.PlaybackPosition()
        log.Printf("Sending Now... playtime:%s \r",playbackPosition)
      }
    }
}

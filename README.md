# discord-tts

text to speech bot for discord.  
(Support CoeFont voice.)

## require

- goenv
- ffmpeg

## run

```bash
$ export TOKEN=XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
$ export COEFONT_ACCESS_TOKEN=xxxxxxxxxxxxxxxxxx
$ export COEFONT_SECRET=xxxxxxxxxxxxxxxxxxxxxxxx
$ go run main.go
```

## usage

1. `@<bot-name> join` : The bot enters the same voice chat as you
2. In same channel of 1 , send caht `hogehuga` : Bot talks to 'hogehuga' in voice chat.

In this sample, the bot says "test".

![sample](./sample.png)

## custom prefix

```
$ go run main.go --prefix=xxx
```

You can use it like this

- `xxx join`
- `xxx leave`

## contribution

Welcome

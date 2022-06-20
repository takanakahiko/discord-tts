# discord-tts

text to speech bot for discord.  
(Support CoeFont voice.)

## require

- ffmpeg

## installation

```bash
$ go get -u github.com/takanakahiko/discord-tts
```

## usage

```bash
$ export TOKEN=XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
$ export COEFONT_ACCESS_TOKEN=xxxxxxxxxxxxxxxxxx
$ export COEFONT_SECRET=xxxxxxxxxxxxxxxxxxxxxxxx
$ discord-tts [--prefix=xxx]
```

## usage

1. `@<bot-name> join` : The bot enters the same voice chat as you
2. In same channel of 1 , send caht `hogehuga` : Bot talks to 'hogehuga' in voice chat.

In this sample, the bot says "test".

![sample](./sample.png)

## custom prefix

```
$ discord-tts --prefix=xxx
```

You can use it like this

- `xxx join`
- `xxx leave`

## debug

```bash
$ export TOKEN=XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
$ export COEFONT_ACCESS_TOKEN=xxxxxxxxxxxxxxxxxx
$ export COEFONT_SECRET=xxxxxxxxxxxxxxxxxxxxxxxx
$ go run cmd/discord-tts/discord-tts.go
```

## contribution

Welcome

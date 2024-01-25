module github.com/takanakahiko/discord-tts

go 1.18

// patch: https://github.com/takanakahiko-fork/dca/commit/7f9f6fb86b20cc6a4af1c95e4c442fba8e4f274d
replace github.com/jonas747/dca v0.0.0-20201113050843-65838623978b => github.com/takanakahiko-fork/dca v0.0.0-20240125163404-7f9f6fb86b20

require (
	cloud.google.com/go/texttospeech v1.6.0
	github.com/bwmarrin/discordgo v0.27.1
	github.com/google/uuid v1.3.0
	github.com/jonas747/dca v0.0.0-20201113050843-65838623978b
	golang.org/x/text v0.14.0
)

require (
	cloud.google.com/go v0.110.0 // indirect
	cloud.google.com/go/compute v1.19.1 // indirect
	cloud.google.com/go/compute/metadata v0.2.3 // indirect
	cloud.google.com/go/longrunning v0.4.1 // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.2.3 // indirect
	github.com/googleapis/gax-go/v2 v2.7.1 // indirect
	github.com/gorilla/websocket v1.5.1 // indirect
	github.com/jonas747/ogg v0.0.0-20161220051205-b4f6f4cf3757 // indirect
	go.opencensus.io v0.24.0 // indirect
	golang.org/x/crypto v0.18.0 // indirect
	golang.org/x/net v0.20.0 // indirect
	golang.org/x/oauth2 v0.7.0 // indirect
	golang.org/x/sys v0.16.0 // indirect
	google.golang.org/api v0.114.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20230410155749-daa745c078e1 // indirect
	google.golang.org/grpc v1.56.3 // indirect
	google.golang.org/protobuf v1.30.0 // indirect
)

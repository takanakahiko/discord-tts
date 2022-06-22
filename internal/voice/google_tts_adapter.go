package voice

import (
	"context"
	"io/ioutil"
	"log"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	texttospeechpb "google.golang.org/genproto/googleapis/cloud/texttospeech/v1"
)

var _ VoiceAdapter = &googleTtsAdapter{}

// googleTtsAdapter
type googleTtsAdapter struct {
	LanguageCode string
}

func NewGoogleTtsAdapter(languageCode string) VoiceAdapter {
	return &googleTtsAdapter{
		LanguageCode: languageCode,
	}
}

func (a *googleTtsAdapter) FetchVoiceUrl(text string) string {
	ctx := context.Background()

	client, err := texttospeech.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	req := texttospeechpb.SynthesizeSpeechRequest{
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{Text: text},
		},
		Voice: nil,
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding: texttospeechpb.AudioEncoding_MP3,
		},
	}

	switch a.LanguageCode {
	case "ja-JP":
		req.Voice = &texttospeechpb.VoiceSelectionParams{
			LanguageCode: a.LanguageCode,
			SsmlGender:   texttospeechpb.SsmlVoiceGender_FEMALE,
			Name:         "ja-JP-Wavenet-B",
		}
	case "en-US":
		req.Voice = &texttospeechpb.VoiceSelectionParams{
			LanguageCode: a.LanguageCode,
			SsmlGender:   texttospeechpb.SsmlVoiceGender_FEMALE,
			Name:         "en-US-Wavenet-C",
		}
	}

	resp, err := client.SynthesizeSpeech(ctx, &req)
	if err != nil {
		log.Fatal(err)
	}

	tmpfile, err := ioutil.TempFile("", "discord-tts_google-tts-adapter_*.mp3")
	if err != nil {
		log.Fatal(err)
	}
	defer tmpfile.Close()
	err = ioutil.WriteFile(tmpfile.Name(), resp.AudioContent, 0644)

	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Audio content written to file: %v\n", tmpfile.Name())
	return tmpfile.Name()
}

package voice

import (
	"context"
	"log"
	"os"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
)

var _ Adapter = (*googleTtsAdapter)(nil)

// googleTtsAdapter.
type googleTtsAdapter struct {
	LanguageCode string
}

func NewGoogleTtsAdapter(languageCode string) *googleTtsAdapter {
	return &googleTtsAdapter{
		LanguageCode: languageCode,
	}
}

func (a *googleTtsAdapter) FetchVoiceURL(text string) string {
	ctx := context.Background()

	client, err := texttospeech.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	req := texttospeechpb.SynthesizeSpeechRequest{
		Input: &texttospeechpb.SynthesisInput{
			InputSource:          &texttospeechpb.SynthesisInput_Text{Text: text},
			Prompt:               nil,
			CustomPronunciations: nil,
		},
		Voice: nil,
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding: texttospeechpb.AudioEncoding_MP3,

			// 以下デフォルト値
			SpeakingRate:     0,
			Pitch:            0,
			VolumeGainDb:     0,
			SampleRateHertz:  0,
			EffectsProfileId: nil,
		},
		AdvancedVoiceOptions: nil,
	}

	switch a.LanguageCode {
	case "ja-JP":
		req.Voice = &texttospeechpb.VoiceSelectionParams{
			LanguageCode:            a.LanguageCode,
			SsmlGender:              texttospeechpb.SsmlVoiceGender_FEMALE,
			Name:                    "ja-JP-Wavenet-B",
			CustomVoice:             nil,
			VoiceClone:              nil,
			ModelName:               "",
			MultiSpeakerVoiceConfig: nil,
		}
	case "en-US":
		req.Voice = &texttospeechpb.VoiceSelectionParams{
			LanguageCode:            a.LanguageCode,
			SsmlGender:              texttospeechpb.SsmlVoiceGender_FEMALE,
			Name:                    "en-US-Wavenet-C",
			CustomVoice:             nil,
			VoiceClone:              nil,
			ModelName:               "",
			MultiSpeakerVoiceConfig: nil,
		}
	}

	resp, err := client.SynthesizeSpeech(ctx, &req)
	if err != nil {
		log.Panic(err)
	}

	tmpfile, err := os.CreateTemp("", "discord-tts_google-tts-adapter_*.mp3")
	if err != nil {
		log.Panic(err)
	}
	defer tmpfile.Close()
	err = os.WriteFile(tmpfile.Name(), resp.GetAudioContent(), 0600) //nolint:gofumpt,mnd // fs.FileModeは直接指定した方がわかりやすいため

	if err != nil {
		log.Panic(err)
	}
	log.Printf("Audio content written to file: %v\n", tmpfile.Name())
	return tmpfile.Name()
}

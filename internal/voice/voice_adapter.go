package voice

type VoiceAdapter interface {
	FetchVoiceUrl(text string) string
}

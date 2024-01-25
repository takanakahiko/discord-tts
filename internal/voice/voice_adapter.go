package voice

type Adapter interface {
	FetchVoiceURL(text string) string
}

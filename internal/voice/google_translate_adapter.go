package voice

import (
	"fmt"
	"net/url"
)

var _ Adapter = &googleTranslateAdapter{}

// googleTranslateAdapter
// 仮の実装として使っているが本来は利用しないほうがいい。
type googleTranslateAdapter struct {
	Lang string
}

func NewGoogleTranslateAdapter(lang string) *googleTranslateAdapter {
	return &googleTranslateAdapter{Lang: lang}
}

func (a *googleTranslateAdapter) FetchVoiceURL(text string) string {
	return fmt.Sprintf(
		"http://translate.google.com/translate_tts?ie=UTF-8&textlen=32&client=tw-ob&q=%s&tl=%s",
		url.QueryEscape(text), a.Lang)
}

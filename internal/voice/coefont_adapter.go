package voice

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
)

var _ Adapter = (*coefontAdapter)(nil)

type coefontAdapter struct {
	CoefontID string
}

func NewCoefontAdapter(coefontID string) *coefontAdapter {
	return &coefontAdapter{CoefontID: coefontID}
}

type text2SpeechReq struct {
	CoefontID string  `json:"coefont,omitempty"`
	Text      string  `json:"text,omitempty"`
	Speed     float64 `json:"speed"`
}

func (a *coefontAdapter) FetchVoiceURL(text string) string {
	ctx := context.Background()

	accessKey := os.Getenv("COEFONT_ACCESS_TOKEN")
	secret := os.Getenv("COEFONT_SECRET")

	bytejson, err := json.Marshal(text2SpeechReq{
		CoefontID: a.CoefontID,
		Text:      text,
		Speed:     0.7, //nolint:mnd // 直接指定した方がコードの可読性が高いため
	})
	if err != nil {
		return ""
	}
	stringtime := strconv.FormatInt(time.Now().Unix(), 10)
	sign := calcHMACSHA256(stringtime+string(bytejson), secret)

	client := &http.Client{
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},

		// 以下デフォルト値
		Transport: nil,
		Jar:       nil,
		Timeout:   0,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.coefont.cloud/v1/text2speech", bytes.NewBuffer(bytejson))
	if err != nil {
		return ""
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Coefont-Content", sign)
	req.Header.Set("X-Coefont-Date", stringtime)
	req.Header.Set("Authorization", accessKey)
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return ""
	}

	resp2, err := http.NewRequestWithContext(ctx, http.MethodGet, resp.Header.Get("Location"), nil)
	if err != nil {
		return ""
	}
	defer resp2.Body.Close()
	u, err := uuid.NewRandom()
	if err != nil {
		return ""
	}
	uu := u.String()
	path := uu + ".wav"
	audiofile, err := os.Create(path)
	if err != nil {
		return ""
	}
	defer audiofile.Close()
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return ""
	}
	_, err = audiofile.Write(buf.Bytes())
	if err != nil {
		return ""
	}

	currentDirectory, _ := os.Getwd()
	return currentDirectory + "/" + path
}

func calcHMACSHA256(message, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

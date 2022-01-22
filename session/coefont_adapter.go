package session

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/google/uuid"
	"net/http"
	"os"
	"strconv"
	"time"
)

type CoefontAdapter struct {
	CoefontID string
}

func NewCoefontAdapter(coefontID string) *CoefontAdapter {
	return &CoefontAdapter{CoefontID: coefontID}
}

type text2SpeechReq struct {
	CoefontID string  `json:"coefont,omitempty"`
	Text      string  `json:"text,omitempty"`
	Speed     float64 `json:"speed"`
}

func (a *CoefontAdapter) FetchVoiceUrl(text string) string {
	accessKey := os.Getenv("COEFONT_ACCESS_TOKEN")
	secret    := os.Getenv("COEFONT_SECRET")

	j, err := json.Marshal(text2SpeechReq{
		CoefontID: a.CoefontID,
		Text:      text,
		Speed:     0.7,
	})
	if err != nil {
		return ""
	}
	t := strconv.FormatInt(time.Now().Unix(), 10)
	sign := a.calcHMACSHA256(t+string(j), secret)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequest("POST", "https://api.coefont.cloud/v1/text2speech", bytes.NewBuffer(j))
	if err != nil {
		return ""
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Coefont-Content", sign)
	req.Header.Set("X-Coefont-Date", t)
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

	resp, err = http.Get(resp.Header.Get("Location"))
	if err != nil {
		return ""
	}
	u, err := uuid.NewRandom()
	uu := u.String()
	path := uu + ".wav"
	f, err := os.Create(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return ""
	}
	_, err = f.Write(buf.Bytes())
	if err != nil {
		return ""
	}

	currentDirectory, _ := os.Getwd()
	return currentDirectory + "/" + path
}

func (a *CoefontAdapter) calcHMACSHA256(message, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

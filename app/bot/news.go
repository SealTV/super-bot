package bot

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	log "github.com/go-pkgz/lgr"
)

// News bot, returns 5 last articles in MD from https://news.radio-t.com/api/v1/news/lastmd/5
type News struct {
	newsAPI string
}

type newsArticle struct {
	Title string    `json:"title"`
	Link  string    `json:"link"`
	Ts    time.Time `json:"ats"`
}

// NewNews makes new News bot
func NewNews(api string) *News {
	log.Printf("[INFO] news bot with api %s", api)
	return &News{newsAPI: api}
}

// OnMessage returns 5 last news articles
func (n News) OnMessage(msg Message) (response string, answer bool) {
	if !contains(n.ReactOn(), msg.Text) {
		return "", false
	}

	reqURL := fmt.Sprintf("%s/v1/news/last/5", n.newsAPI)
	log.Printf("[DEBUG] request %s", reqURL)
	client := http.Client{Timeout: time.Second * 5}
	resp, err := client.Get(reqURL)
	if err != nil {
		log.Printf("[WARN] failed to send request %s, error=%v", reqURL, err)
		return "", false
	}
	defer resp.Body.Close()

	articles := []newsArticle{}
	if err = json.NewDecoder(resp.Body).Decode(&articles); err != nil {
		log.Printf("[WARN] failed to parse response, error %v", err)
		return "", false
	}

	var lines []string
	for _, a := range articles {
		lines = append(lines, fmt.Sprintf("- [%s](%s) %s", a.Title, a.Link, a.Ts.Format("2006-01-02")))
	}
	return strings.Join(lines, "\n") + "\n- [все новости и темы](https://news.radio-t.com)", true
}

// ReactOn keys
func (n News) ReactOn() []string {
	return []string{"news!", "новости!"}
}

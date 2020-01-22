package bot

import (
	"context"
	"net/http"
	"sync"
	"time"

	log "github.com/go-pkgz/lgr"
)

const (
	MsgBroadcastStarted  = "Вещание началось"
	MsgBroadcastFinished = "Вещание завершилось"
)

type BroadcastParams struct {
	Url          string        // Url for "ping"
	PingInterval time.Duration // Ping interval
	DelayToOff   time.Duration // State will be switched to off in no ok replies from Url in this intrval
	Client       http.Client   // http client
}

// BroadcastStatus bot replies with current broadcast status
type BroadcastStatus struct {
	status         bool // current broadcast status
	lastSentStatus bool // last status sent with OnMessage
	fistMsgSent    bool
	statusMx       sync.Mutex
}

// NewBroadcastStatus starts status checking goroutine and returns bot instance
func NewBroadcastStatus(ctx context.Context, params BroadcastParams) *BroadcastStatus {
	log.Printf("[INFO] BroadcastStatus bot with %v", params.Url)
	b := &BroadcastStatus{}
	go b.checker(ctx, params)
	return b
}

// OnMessage returns current broadcast status if it was changed
func (b *BroadcastStatus) OnMessage(_ Message) (response string, answer bool) {
	b.statusMx.Lock()
	defer b.statusMx.Unlock()

	if b.status == b.lastSentStatus && b.fistMsgSent {
		return
	}

	b.fistMsgSent = true
	answer = true
	response = MsgBroadcastFinished

	b.lastSentStatus = b.status
	if b.status {
		response = MsgBroadcastStarted
	}
	return
}

func (b *BroadcastStatus) checker(ctx context.Context, params BroadcastParams) {
	lastOn := time.Time{}
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(params.PingInterval):
			lastOn = b.check(ctx, lastOn, params)
		}
	}
}

// check do ping to url and change current state
func (b *BroadcastStatus) check(ctx context.Context, lastOn time.Time, params BroadcastParams) time.Time {
	b.statusMx.Lock()
	defer b.statusMx.Unlock()

	newStatus := ping(ctx, params.Client, params.Url)

	// 0 -> 1
	if !b.status && newStatus {
		log.Print("[INFO] Broadcast started")
		b.status = true
		return time.Now()
	}

	// 1 -> 0
	// 0 -> 0
	if !newStatus {
		if b.status && lastOn.Add(params.DelayToOff).Before(time.Now()) {
			log.Print("[INFO] Broadcast finished")
			b.status = false
		}
		return lastOn
	}

	// 1 -> 1
	return time.Now()
}

// ping do get request to https://stream.radio-t.com and returns true on OK status and false for all other statuses
func ping(ctx context.Context, client http.Client, url string) (status bool) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Printf("[WARN] unable to created %v request, %v", url, err)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[WARN] unable to do %v request, %v", url, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		status = true
	}
	return
}

func (b *BroadcastStatus) getStatus() bool {
	b.statusMx.Lock()
	defer b.statusMx.Unlock()
	return b.status
}

// ReactOn keys
func (b *BroadcastStatus) ReactOn() []string {
	return []string{}
}
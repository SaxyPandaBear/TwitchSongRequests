package api_test

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/nicklaw5/helix"
	"github.com/stretchr/testify/assert"
	"github.com/zmb3/spotify/v2"

	handler "github.com/saxypandabear/twitchsongrequests/pkg/api"
	"github.com/saxypandabear/twitchsongrequests/pkg/db"

	"github.com/saxypandabear/twitchsongrequests/pkg/queue"
)

type dummyPublisher struct {
	messages   chan string
	shouldFail bool
}

type mockReadCloser struct{}

func (m mockReadCloser) Read(p []byte) (int, error) {
	return 0, errors.New("expected to fail")
}
func (m mockReadCloser) Close() error {
	return nil
}

type headerTestCase struct {
	header       string
	verification string
	shouldPass   bool
}

var (
	_ queue.Publisher = dummyPublisher{
		messages:   nil,
		shouldFail: false,
	}
	dummySecret         = "dummy"
	eventSubMsgID       = "foo"
	size                = 20
	letters             = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	replace             = "REPLACEME"
	msgIDHeader         = "Twitch-Eventsub-Message-Id"
	msgTimestampHeader  = "Twitch-Eventsub-Message-Timestamp"
	msgSignatureHeader  = "Twitch-Eventsub-Message-Signature"
	testResponseTimeout = time.Millisecond * 50
)

//go:embed testdata/redeem.json
var redeemPayload string

//go:embed testdata/verification.json
var verificationPayload string

func (p dummyPublisher) Publish(client *spotify.Client, url string) error {
	if p.shouldFail {
		return errors.New("oops")
	}

	p.messages <- url
	return nil
}

func TestPublishRedeem(t *testing.T) {
	m := make(chan string)
	p := dummyPublisher{
		messages:   m,
		shouldFail: false,
	}
	u := db.InMemoryUserStore{}

	rh := handler.NewRewardHandler(dummySecret, &p, &u)

	userInput := generateUserInput(t)
	payload := strings.Replace(redeemPayload, replace, userInput, 1)
	assert.NotEmpty(t, payload)
	assert.False(t, strings.Contains(payload, replace))

	var payloadMap handler.EventSubNotification
	err := json.Unmarshal([]byte(payload), &payloadMap)
	assert.NoError(t, err)
	assert.NotNil(t, payloadMap)

	req, err := http.NewRequest("POST", "/callback", strings.NewReader(payload))
	assert.NoError(t, err)

	// spoof signature header
	ts := time.Now().Format(time.RFC3339)
	sig := deriveEventsubSignature(t, payload, eventSubMsgID, ts, dummySecret)
	req.Header.Add(msgIDHeader, eventSubMsgID)
	req.Header.Add(msgTimestampHeader, ts)
	req.Header.Add(msgSignatureHeader, sig)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(rh.ChannelPointRedeem)

	go func() {
		handler.ServeHTTP(rr, req)
	}()

	var event string
	select {
	case event = <-m:
		t.Logf("received %v", event)
	case <-time.After(testResponseTimeout):
		t.Error("did not receive message in time")
	}

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.NotEmpty(t, event)

	assert.Equal(t, userInput, event)
}

func TestPublishRedeemEmptyBody(t *testing.T) {
	m := make(chan string)
	p := dummyPublisher{
		messages:   m,
		shouldFail: false,
	}
	u := db.InMemoryUserStore{}

	rh := handler.NewRewardHandler(dummySecret, &p, &u)

	userInput := generateUserInput(t)
	payload := strings.Replace(redeemPayload, replace, userInput, 1)
	assert.NotEmpty(t, payload)
	assert.False(t, strings.Contains(payload, replace))

	var payloadMap handler.EventSubNotification
	err := json.Unmarshal([]byte(payload), &payloadMap)
	assert.NoError(t, err)
	assert.NotNil(t, payloadMap)

	// pass in a mockReadCloser that always fails on read
	req, err := http.NewRequest("POST", "/callback", &mockReadCloser{})
	assert.NoError(t, err)

	// spoof signature header
	ts := time.Now().Format(time.RFC3339)
	sig := deriveEventsubSignature(t, payload, eventSubMsgID, ts, dummySecret)
	req.Header.Add(msgIDHeader, eventSubMsgID)
	req.Header.Add(msgTimestampHeader, ts)
	req.Header.Add(msgSignatureHeader, sig)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(rh.ChannelPointRedeem)

	go func() {
		handler.ServeHTTP(rr, req)
	}()

	select {
	case <-m:
		t.Error("should not have received a message")
	case <-time.After(testResponseTimeout):
		t.Log("no event expected")
	}

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestPublishRedeemFails(t *testing.T) {
	m := make(chan string)
	p := dummyPublisher{
		messages:   m,
		shouldFail: true,
	}
	u := db.InMemoryUserStore{}

	rh := handler.NewRewardHandler(dummySecret, &p, &u)

	userInput := generateUserInput(t)
	payload := strings.Replace(redeemPayload, replace, userInput, 1)
	assert.NotEmpty(t, payload)
	assert.False(t, strings.Contains(payload, replace))

	var payloadMap handler.EventSubNotification
	err := json.Unmarshal([]byte(payload), &payloadMap)
	assert.NoError(t, err)
	assert.NotNil(t, payloadMap)

	req, err := http.NewRequest("POST", "/callback", strings.NewReader(payload))
	assert.NoError(t, err)

	// spoof signature header
	ts := time.Now().Format(time.RFC3339)
	sig := deriveEventsubSignature(t, payload, eventSubMsgID, ts, dummySecret)
	req.Header.Add(msgIDHeader, eventSubMsgID)
	req.Header.Add(msgTimestampHeader, ts)
	req.Header.Add(msgSignatureHeader, sig)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(rh.ChannelPointRedeem)

	go func() {
		handler.ServeHTTP(rr, req)
	}()

	select {
	case <-m:
		t.Error("should not have received a message")
	case <-time.After(testResponseTimeout):
		t.Log("no event expected")
	}

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestPublishRedeemInvalidSignature(t *testing.T) {
	m := make(chan string)
	p := dummyPublisher{
		messages:   m,
		shouldFail: false,
	}
	u := db.InMemoryUserStore{}

	rh := handler.NewRewardHandler(dummySecret, &p, &u)

	userInput := generateUserInput(t)
	payload := strings.Replace(redeemPayload, replace, userInput, 1)
	assert.NotEmpty(t, payload)
	assert.False(t, strings.Contains(payload, replace))

	req, err := http.NewRequest("POST", "/callback", strings.NewReader(payload))
	assert.NoError(t, err)

	// spoof signature header
	ts := time.Now().Format(time.RFC3339)
	sig := deriveEventsubSignature(t, payload, eventSubMsgID, ts, dummySecret)
	req.Header.Add(msgIDHeader, eventSubMsgID)
	req.Header.Add(msgTimestampHeader, ts)
	req.Header.Add(msgSignatureHeader, sig+"abc123") // signature header exists, but invalid

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(rh.ChannelPointRedeem)

	go func() {
		handler.ServeHTTP(rr, req)
	}()

	select {
	case <-m:
		t.Error("should not have received a message")
	case <-time.After(testResponseTimeout):
		t.Log("no event expected")
	}

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestPublishRedeemInvalidJSON(t *testing.T) {
	m := make(chan string)
	p := dummyPublisher{
		messages:   m,
		shouldFail: false,
	}
	u := db.InMemoryUserStore{}

	rh := handler.NewRewardHandler(dummySecret, &p, &u)

	userInput := generateUserInput(t)
	payload := strings.Replace(redeemPayload, replace, userInput, 1)
	payload = strings.Replace(payload, "}", "foo", -1) // should be invalid json now
	assert.NotEmpty(t, payload)
	assert.False(t, strings.Contains(payload, replace))

	var payloadMap handler.EventSubNotification
	err := json.Unmarshal([]byte(payload), &payloadMap)
	assert.Error(t, err)

	req, err := http.NewRequest("POST", "/callback", strings.NewReader(payload))
	assert.NoError(t, err)

	// spoof signature header
	ts := time.Now().Format(time.RFC3339)
	sig := deriveEventsubSignature(t, payload, eventSubMsgID, ts, dummySecret)
	req.Header.Add(msgIDHeader, eventSubMsgID)
	req.Header.Add(msgTimestampHeader, ts)
	req.Header.Add(msgSignatureHeader, sig)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(rh.ChannelPointRedeem)

	go func() {
		handler.ServeHTTP(rr, req)
	}()

	select {
	case <-m:
		t.Error("should not have received a message")
	case <-time.After(testResponseTimeout):
		t.Log("no event expected")
	}

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestPublishRedeemInvalidPayload(t *testing.T) {
	m := make(chan string)
	p := dummyPublisher{
		messages:   m,
		shouldFail: false,
	}
	u := db.InMemoryUserStore{}

	rh := handler.NewRewardHandler(dummySecret, &p, &u)

	userInput := generateUserInput(t)
	payload := strings.Replace(redeemPayload, replace, userInput, 1)
	payload = strings.Replace(payload, "\"broadcaster_user_id\": \"12826\"", "\"broadcaster_user_id\": 12826", 1)
	assert.NotEmpty(t, payload)
	assert.False(t, strings.Contains(payload, replace))

	var payloadMap handler.EventSubNotification
	err := json.Unmarshal([]byte(payload), &payloadMap)
	assert.NoError(t, err)
	assert.NotNil(t, payloadMap)
	var redeemEvent helix.EventSubChannelPointsCustomRewardRedemptionEvent
	err = json.NewDecoder(bytes.NewReader(payloadMap.Event)).Decode(&redeemEvent)
	assert.Error(t, err)

	req, err := http.NewRequest("POST", "/callback", strings.NewReader(payload))
	assert.NoError(t, err)

	// spoof signature header
	ts := time.Now().Format(time.RFC3339)
	sig := deriveEventsubSignature(t, payload, eventSubMsgID, ts, dummySecret)
	req.Header.Add(msgIDHeader, eventSubMsgID)
	req.Header.Add(msgTimestampHeader, ts)
	req.Header.Add(msgSignatureHeader, sig)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(rh.ChannelPointRedeem)

	go func() {
		handler.ServeHTTP(rr, req)
	}()

	select {
	case <-m:
		t.Error("should not have received a message")
	case <-time.After(testResponseTimeout):
		t.Log("no event expected")
	}

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// The endpoint used for webhook callbacks must also verify itself:
// https://dev.twitch.tv/docs/eventsub/handling-webhook-events/#responding-to-a-challenge-request
func TestVerifyWebhookCallback(t *testing.T) {
	m := make(chan string)
	p := dummyPublisher{
		messages:   m,
		shouldFail: false,
	}
	u := db.InMemoryUserStore{}

	rh := handler.NewRewardHandler(dummySecret, &p, &u)

	challenge := generateUserInput(t)
	payload := strings.Replace(verificationPayload, replace, challenge, 1)
	assert.NotEmpty(t, payload)
	assert.False(t, strings.Contains(payload, replace))

	var payloadMap handler.EventSubNotification
	err := json.Unmarshal([]byte(payload), &payloadMap)
	assert.NoError(t, err)
	assert.NotNil(t, payloadMap)

	req, err := http.NewRequest("POST", "/callback", strings.NewReader(payload))
	assert.NoError(t, err)

	// spoof signature header
	ts := time.Now().Format(time.RFC3339)
	sig := deriveEventsubSignature(t, payload, eventSubMsgID, ts, dummySecret)
	req.Header.Add(msgIDHeader, eventSubMsgID)
	req.Header.Add(msgTimestampHeader, ts)
	req.Header.Add(msgSignatureHeader, sig)

	// add header so that the service knows that Twitch is trying to verify the callback
	req.Header.Add("Twitch-Eventsub-Message-Type", "webhook_callback_verification")

	assert.True(t, handler.IsVerificationRequest(req))

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(rh.ChannelPointRedeem)

	go func() {
		handler.ServeHTTP(rr, req)
	}()

	select {
	case <-m:
		t.Error("should not have received a message")
	case <-time.After(testResponseTimeout):
		t.Log("no event expected")
	}

	assert.Equal(t, challenge, rr.Body.String())
}

func TestIsVerificationRequest(t *testing.T) {
	tests := []headerTestCase{
		{
			header:       "foo",
			verification: "bar",
		},
		{
			header:       "Twitch-Eventsub-Message-Type",
			verification: "webhook_callback_verification",
			shouldPass:   true,
		},
		{
			verification: "foo",
		},
		{
			header: "foo",
		},
		{},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s: %s", test.header, test.verification), func(t *testing.T) {
			req, err := http.NewRequest("POST", "/foo", strings.NewReader("hello, world!"))
			assert.NoError(t, err)

			req.Header.Add(test.header, test.verification)

			assert.Equal(t, test.shouldPass, handler.IsVerificationRequest(req))
		})
	}
}

func generateUserInput(t *testing.T) string {
	t.Helper()
	rand.Seed(time.Now().UnixNano())
	pseudo := make([]rune, size)
	for i := range pseudo {
		pseudo[i] = letters[rand.Intn(len(letters))]
	}
	return fmt.Sprintf("%s:%s", time.Now().Format(time.RFC3339), string(pseudo))
}

func deriveEventsubSignature(t *testing.T, payload, messageID, timestamp, secret string) string {
	t.Helper()
	hmacMessage := []byte(fmt.Sprintf("%s%s%s", messageID, timestamp, payload))
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(hmacMessage)
	return fmt.Sprintf("sha256=%s", hex.EncodeToString(mac.Sum(nil)))
}

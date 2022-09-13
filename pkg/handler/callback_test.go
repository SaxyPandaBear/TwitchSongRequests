package handler_test

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

	"github.com/saxypandabear/twitchsongrequests/pkg/handler"

	"github.com/saxypandabear/twitchsongrequests/pkg/queue"
)

type dummyPublisher struct {
	messages   chan interface{}
	shouldFail bool
}

type mockReadCloser struct{}

func (m mockReadCloser) Read(p []byte) (int, error) {
	return 0, errors.New("expected to fail")
}
func (m mockReadCloser) Close() error {
	return nil
}

// make sure that dummyPublisher maintains the Publisher interface
var (
	_ queue.Publisher = dummyPublisher{
		messages:   nil,
		shouldFail: false,
	}
	dummySecret        = "dummy"
	eventSubMsgID      = "foo"
	size               = 20
	letters            = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	replace            = "REPLACEME"
	msgIDHeader        = "Twitch-Eventsub-Message-Id"
	msgTimestampHeader = "Twitch-Eventsub-Message-Timestamp"
	msgSignatureHeader = "Twitch-Eventsub-Message-Signature"
)

//go:embed testdata/redeem.json
var redeemPayload string

func (p dummyPublisher) Publish(val interface{}) error {
	if p.shouldFail {
		return errors.New("oops")
	}

	p.messages <- val
	return nil
}

func TestPublishRedeem(t *testing.T) {
	m := make(chan interface{})
	p := dummyPublisher{
		messages:   m,
		shouldFail: false,
	}

	rh := handler.NewRewardHandler(dummySecret, p)

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

	var event interface{}
	select {
	case event = <-m:
		t.Logf("received %v", event)
	case <-time.After(time.Second):
		t.Error("did not receive message in time")
	}

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.NotNil(t, event)
	eventMsg, ok := event.(string)
	assert.True(t, ok)
	assert.Equal(t, userInput, eventMsg)
}

func TestPublishRedeemEmptyBody(t *testing.T) {
	m := make(chan interface{})
	p := dummyPublisher{
		messages:   m,
		shouldFail: false,
	}

	rh := handler.NewRewardHandler(dummySecret, p)

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

	var event interface{}
	select {
	case event = <-m:
		t.Error("should not have received a message")
	case <-time.After(time.Second):
		t.Log("no event expected")
	}

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Nil(t, event)
}

func TestPublishRedeemFails(t *testing.T) {
	m := make(chan interface{})
	p := dummyPublisher{
		messages:   m,
		shouldFail: true,
	}

	rh := handler.NewRewardHandler(dummySecret, p)

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

	var event interface{}
	select {
	case event = <-m:
		t.Error("should not have received a message")
	case <-time.After(time.Second):
		t.Log("no event expected")
	}

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Nil(t, event)
}

func TestPublishRedeemInvalidSignature(t *testing.T) {
	m := make(chan interface{})
	p := dummyPublisher{
		messages:   m,
		shouldFail: false,
	}

	rh := handler.NewRewardHandler(dummySecret, p)

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

	var event interface{}
	select {
	case event = <-m:
		t.Error("should not have received a message")
	case <-time.After(time.Second):
		t.Log("no event expected")
	}

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Nil(t, event)
}

func TestPublishRedeemInvalidJSON(t *testing.T) {
	m := make(chan interface{})
	p := dummyPublisher{
		messages:   m,
		shouldFail: false,
	}

	rh := handler.NewRewardHandler(dummySecret, p)

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

	var event interface{}
	select {
	case event = <-m:
		t.Error("should not have received a message")
	case <-time.After(time.Second):
		t.Log("no event expected")
	}

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Nil(t, event)
}

func TestPublishRedeemInvalidPayload(t *testing.T) {
	m := make(chan interface{})
	p := dummyPublisher{
		messages:   m,
		shouldFail: false,
	}

	rh := handler.NewRewardHandler(dummySecret, p)

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

	var event interface{}
	select {
	case event = <-m:
		t.Error("should not have received a message")
	case <-time.After(time.Second):
		t.Log("no event expected")
	}

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Nil(t, event)
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

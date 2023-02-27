package api_test

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/nicklaw5/helix"
	"github.com/stretchr/testify/assert"

	"github.com/saxypandabear/twitchsongrequests/internal/testutil"
	"github.com/saxypandabear/twitchsongrequests/internal/util"
	"github.com/saxypandabear/twitchsongrequests/pkg/api"
	"github.com/saxypandabear/twitchsongrequests/pkg/users"
)

var (
	dummySecret            = "dummy"
	eventSubMsgID          = "foo"
	size                   = 20
	letters                = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	userInputPlaceholder   = "TESTUSERINPUT"
	challengePlaceholder   = "CHALLENGEINPUT"
	rewardTitlePlaceholder = "TESTREWARDTITLE"
	msgIDHeader            = "Twitch-Eventsub-Message-Id"
	msgTimestampHeader     = "Twitch-Eventsub-Message-Timestamp"
	msgSignatureHeader     = "Twitch-Eventsub-Message-Signature"
	testResponseTimeout    = time.Millisecond * 50
)

//go:embed testdata/redeem.json
var redeemPayload string

//go:embed testdata/verification.json
var verificationPayload string

func TestPublishRedeem(t *testing.T) {
	m := make(chan string)
	p := testutil.DummyPublisher{
		Messages:   m,
		ShouldFail: false,
	}
	u := testutil.InMemoryUserStore{
		Data: make(map[string]*users.User),
	}
	err := u.AddUser(&users.User{ // spoof a user so the test doesen't fail
		TwitchID:            "12826",
		SpotifyAccessToken:  "foo",
		SpotifyRefreshToken: "bar",
		SpotifyExpiry:       &time.Time{},
	})
	assert.NoError(t, err)

	rh := api.NewRewardHandler(dummySecret, &p, &u, &util.AuthConfig{})

	userInput := generateUserInput(t)
	payload := strings.Replace(redeemPayload, userInputPlaceholder, userInput, 1)
	payload = strings.Replace(payload, rewardTitlePlaceholder, api.SongRequestsTitle, 1)
	assert.NotEmpty(t, payload)
	assert.False(t, strings.Contains(payload, userInputPlaceholder))

	var payloadMap api.EventSubNotification
	err = json.Unmarshal([]byte(payload), &payloadMap)
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
	api := http.HandlerFunc(rh.ChannelPointRedeem)

	go func() {
		api.ServeHTTP(rr, req)
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
	p := testutil.DummyPublisher{
		Messages:   m,
		ShouldFail: false,
	}
	u := testutil.InMemoryUserStore{
		Data: make(map[string]*users.User),
	}

	rh := api.NewRewardHandler(dummySecret, &p, &u, &util.AuthConfig{})

	userInput := generateUserInput(t)
	payload := strings.Replace(redeemPayload, userInputPlaceholder, userInput, 1)
	assert.NotEmpty(t, payload)
	assert.False(t, strings.Contains(payload, userInputPlaceholder))

	var payloadMap api.EventSubNotification
	err := json.Unmarshal([]byte(payload), &payloadMap)
	assert.NoError(t, err)
	assert.NotNil(t, payloadMap)

	// pass in a testutil.MockReadCloser that always fails on read
	req, err := http.NewRequest("POST", "/callback", &testutil.MockReadCloser{})
	assert.NoError(t, err)

	// spoof signature header
	ts := time.Now().Format(time.RFC3339)
	sig := deriveEventsubSignature(t, payload, eventSubMsgID, ts, dummySecret)
	req.Header.Add(msgIDHeader, eventSubMsgID)
	req.Header.Add(msgTimestampHeader, ts)
	req.Header.Add(msgSignatureHeader, sig)

	rr := httptest.NewRecorder()
	api := http.HandlerFunc(rh.ChannelPointRedeem)

	go func() {
		api.ServeHTTP(rr, req)
	}()

	select {
	case <-m:
		t.Error("should not have received a message")
	case <-time.After(testResponseTimeout):
		t.Log("no event expected")
	}

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestPublishIncorrectRewardTitle(t *testing.T) {
	m := make(chan string)
	p := testutil.DummyPublisher{
		Messages:   m,
		ShouldFail: true,
	}
	u := testutil.InMemoryUserStore{
		Data: make(map[string]*users.User),
	}

	rh := api.NewRewardHandler(dummySecret, &p, &u, &util.AuthConfig{})

	userInput := generateUserInput(t)
	payload := strings.Replace(redeemPayload, userInputPlaceholder, userInput, 1)
	assert.NotEmpty(t, payload)
	assert.False(t, strings.Contains(payload, userInputPlaceholder))
	assert.False(t, strings.Contains(payload, api.SongRequestsTitle))

	var payloadMap api.EventSubNotification
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
	api := http.HandlerFunc(rh.ChannelPointRedeem)

	go func() {
		api.ServeHTTP(rr, req)
	}()

	select {
	case <-m:
		t.Error("should not have received a message")
	case <-time.After(testResponseTimeout):
		t.Log("no event expected")
	}

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestPublishNoAuthenticatedUser(t *testing.T) {
	m := make(chan string)
	p := testutil.DummyPublisher{
		Messages:   m,
		ShouldFail: true,
	}
	// the user lookup will fail
	u := testutil.InMemoryUserStore{
		Data: make(map[string]*users.User),
	}

	rh := api.NewRewardHandler(dummySecret, &p, &u, &util.AuthConfig{})

	userInput := generateUserInput(t)
	payload := strings.Replace(redeemPayload, userInputPlaceholder, userInput, 1)
	payload = strings.Replace(payload, rewardTitlePlaceholder, api.SongRequestsTitle, 1)
	assert.NotEmpty(t, payload)
	assert.False(t, strings.Contains(payload, userInputPlaceholder))

	var payloadMap api.EventSubNotification
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
	api := http.HandlerFunc(rh.ChannelPointRedeem)

	go func() {
		api.ServeHTTP(rr, req)
	}()

	select {
	case <-m:
		t.Error("should not have received a message")
	case <-time.After(testResponseTimeout):
		t.Log("no event expected")
	}

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestPublishRedeemFails(t *testing.T) {
	m := make(chan string)
	p := testutil.DummyPublisher{
		Messages:   m,
		ShouldFail: true,
	}
	u := testutil.InMemoryUserStore{
		Data: make(map[string]*users.User),
	}
	err := u.AddUser(&users.User{
		TwitchID:            "12826",
		SpotifyAccessToken:  "foo",
		SpotifyRefreshToken: "bar",
		SpotifyExpiry:       &time.Time{},
	})
	assert.NoError(t, err)

	rh := api.NewRewardHandler(dummySecret, &p, &u, &util.AuthConfig{})

	userInput := generateUserInput(t)
	payload := strings.Replace(redeemPayload, userInputPlaceholder, userInput, 1)
	payload = strings.Replace(payload, rewardTitlePlaceholder, api.SongRequestsTitle, 1)
	assert.NotEmpty(t, payload)
	assert.False(t, strings.Contains(payload, userInputPlaceholder))

	var payloadMap api.EventSubNotification
	err = json.Unmarshal([]byte(payload), &payloadMap)
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
	api := http.HandlerFunc(rh.ChannelPointRedeem)

	go func() {
		api.ServeHTTP(rr, req)
	}()

	select {
	case <-m:
		t.Error("should not have received a message")
	case <-time.After(testResponseTimeout):
		t.Log("no event expected")
	}

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestPublishRedeemInvalidSignature(t *testing.T) {
	m := make(chan string)
	p := testutil.DummyPublisher{
		Messages:   m,
		ShouldFail: false,
	}
	u := testutil.InMemoryUserStore{
		Data: make(map[string]*users.User),
	}

	rh := api.NewRewardHandler(dummySecret, &p, &u, &util.AuthConfig{})

	userInput := generateUserInput(t)
	payload := strings.Replace(redeemPayload, userInputPlaceholder, userInput, 1)
	assert.NotEmpty(t, payload)
	assert.False(t, strings.Contains(payload, userInputPlaceholder))

	req, err := http.NewRequest("POST", "/callback", strings.NewReader(payload))
	assert.NoError(t, err)

	// spoof signature header
	ts := time.Now().Format(time.RFC3339)
	sig := deriveEventsubSignature(t, payload, eventSubMsgID, ts, dummySecret)
	req.Header.Add(msgIDHeader, eventSubMsgID)
	req.Header.Add(msgTimestampHeader, ts)
	req.Header.Add(msgSignatureHeader, sig+"abc123") // signature header exists, but invalid

	rr := httptest.NewRecorder()
	api := http.HandlerFunc(rh.ChannelPointRedeem)

	go func() {
		api.ServeHTTP(rr, req)
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
	p := testutil.DummyPublisher{
		Messages:   m,
		ShouldFail: false,
	}
	u := testutil.InMemoryUserStore{
		Data: make(map[string]*users.User),
	}

	rh := api.NewRewardHandler(dummySecret, &p, &u, &util.AuthConfig{})

	userInput := generateUserInput(t)
	payload := strings.Replace(redeemPayload, userInputPlaceholder, userInput, 1)
	payload = strings.Replace(payload, "}", "foo", -1) // should be invalid json now
	assert.NotEmpty(t, payload)
	assert.False(t, strings.Contains(payload, userInputPlaceholder))

	var payloadMap api.EventSubNotification
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
	api := http.HandlerFunc(rh.ChannelPointRedeem)

	go func() {
		api.ServeHTTP(rr, req)
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
	p := testutil.DummyPublisher{
		Messages:   m,
		ShouldFail: false,
	}
	u := testutil.InMemoryUserStore{
		Data: make(map[string]*users.User),
	}

	rh := api.NewRewardHandler(dummySecret, &p, &u, &util.AuthConfig{})

	userInput := generateUserInput(t)
	payload := strings.Replace(redeemPayload, userInputPlaceholder, userInput, 1)
	payload = strings.Replace(payload, "\"broadcaster_user_id\": \"12826\"", "\"broadcaster_user_id\": 12826", 1)
	assert.NotEmpty(t, payload)
	assert.False(t, strings.Contains(payload, userInputPlaceholder))

	var payloadMap api.EventSubNotification
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
	api := http.HandlerFunc(rh.ChannelPointRedeem)

	go func() {
		api.ServeHTTP(rr, req)
	}()

	select {
	case <-m:
		t.Error("should not have received a message")
	case <-time.After(testResponseTimeout):
		t.Log("no event expected")
	}

	assert.Equal(t, http.StatusOK, rr.Code)
}

// The endpoint used for webhook callbacks must also verify itself:
// https://dev.twitch.tv/docs/eventsub/handling-webhook-events/#responding-to-a-challenge-request
func TestVerifyWebhookCallback(t *testing.T) {
	m := make(chan string)
	p := testutil.DummyPublisher{
		Messages:   m,
		ShouldFail: false,
	}
	u := testutil.InMemoryUserStore{
		Data: make(map[string]*users.User),
	}

	rh := api.NewRewardHandler(dummySecret, &p, &u, &util.AuthConfig{})

	challenge := generateUserInput(t)
	payload := strings.Replace(verificationPayload, challengePlaceholder, challenge, 1)
	assert.NotEmpty(t, payload)
	assert.False(t, strings.Contains(payload, userInputPlaceholder))

	var payloadMap api.EventSubNotification
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

	assert.True(t, api.IsVerificationRequest(req))

	rr := httptest.NewRecorder()
	api := http.HandlerFunc(rh.ChannelPointRedeem)

	go func() {
		api.ServeHTTP(rr, req)
	}()

	select {
	case <-m:
		t.Error("should not have received a message")
	case <-time.After(testResponseTimeout):
		t.Log("no event expected")
	}

	assert.Equal(t, challenge, rr.Body.String())
}

func TestSubscriptionRevoked(t *testing.T) {
	m := make(chan string)
	p := testutil.DummyPublisher{
		Messages:   m,
		ShouldFail: false,
	}
	u := testutil.InMemoryUserStore{
		Data: make(map[string]*users.User),
	}

	rh := api.NewRewardHandler(dummySecret, &p, &u, &util.AuthConfig{})

	challenge := generateUserInput(t)
	payload := strings.Replace(verificationPayload, challengePlaceholder, challenge, 1)
	assert.NotEmpty(t, payload)
	assert.False(t, strings.Contains(payload, userInputPlaceholder))

	var payloadMap api.EventSubNotification
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
	req.Header.Add("Twitch-Eventsub-Message-Type", "revocation")

	assert.True(t, api.IsRevocationRequest(req))

	rr := httptest.NewRecorder()
	api := http.HandlerFunc(rh.ChannelPointRedeem)

	go func() {
		api.ServeHTTP(rr, req)
	}()

	select {
	case <-m:
		t.Error("should not have received a message")
	case <-time.After(testResponseTimeout):
		t.Log("no event expected")
	}
}

func TestIsVerificationRequest(t *testing.T) {
	tests := []struct {
		header       string
		verification string
		shouldPass   bool
	}{
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

			assert.Equal(t, test.shouldPass, api.IsVerificationRequest(req))
		})
	}
}

func TestIsValidSongRequest(t *testing.T) {
	assert.False(t, api.IsValidReward(nil))

	e := helix.EventSubChannelPointsCustomRewardRedemptionEvent{
		Reward: helix.EventSubReward{
			Title: "something",
		},
	}
	assert.False(t, api.IsValidReward(&e))

	e.Reward.Title = fmt.Sprintf("Middle of %s the title", api.SongRequestsTitle)
	assert.True(t, api.IsValidReward(&e))
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

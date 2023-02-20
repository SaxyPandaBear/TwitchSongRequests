package testutil

import (
	"context"
	"errors"
	"net/http"

	"github.com/saxypandabear/twitchsongrequests/pkg/api"
	"github.com/saxypandabear/twitchsongrequests/pkg/queue"
	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

type DummyPublisher struct {
	Messages   chan string
	ShouldFail bool
}

var _ queue.Publisher = (*DummyPublisher)(nil)

func (p DummyPublisher) Publish(client queue.Queuer, url string) error {
	if p.ShouldFail {
		return errors.New("oops")
	}

	p.Messages <- url
	return nil
}

type MockReadCloser struct{}

func (m MockReadCloser) Read(p []byte) (int, error) {
	return 0, errors.New("expected to fail")
}
func (m MockReadCloser) Close() error {
	return nil
}

type MockAuthenticator struct{}

var _ api.IAuthenticator = (*MockAuthenticator)(nil)

func (m MockAuthenticator) Client(ctx context.Context, token *oauth2.Token) *http.Client {
	return http.DefaultClient
}

type MockQueuer struct {
	ShouldFail bool
	Messages   []spotify.ID
}

func (m *MockQueuer) QueueSong(ctx context.Context, trackID spotify.ID) error {
	if m.ShouldFail {
		return errors.New("expected to fail")
	}

	m.Messages = append(m.Messages, trackID)
	return nil
}

package testutil

import (
	"context"
	"errors"

	"github.com/saxypandabear/twitchsongrequests/pkg/queue"
	"github.com/zmb3/spotify/v2"
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

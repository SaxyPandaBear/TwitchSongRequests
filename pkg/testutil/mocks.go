package testutil

import (
	"context"
	"errors"
	"net/http"

	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

type DummyPublisher struct {
	Messages   chan string
	ShouldFail bool
}

func (p DummyPublisher) Publish(client *spotify.Client, url string) error {
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

func (m MockAuthenticator) Client(ctx context.Context, token *oauth2.Token) *http.Client {
	return http.DefaultClient
}

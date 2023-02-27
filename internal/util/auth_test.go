package util_test

import (
	"fmt"
	"testing"

	"github.com/saxypandabear/twitchsongrequests/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestGenerateAuthURL(t *testing.T) {
	c := util.AuthConfig{
		ClientID:    "foo",
		RedirectURL: "bar",
		State:       "baz",
		Scope:       "scope",
	}

	expected := fmt.Sprintf("https://some-host.com/path?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&state=%s",
		c.ClientID, c.RedirectURL, c.Scope, c.State)

	actual := util.GenerateAuthURL("some-host.com", "/path", &c)
	assert.Equal(t, expected, actual)
}

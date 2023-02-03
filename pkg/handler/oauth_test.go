package handler_test

import (
	"fmt"
	"testing"

	"github.com/saxypandabear/twitchsongrequests/pkg/handler"
	"github.com/stretchr/testify/assert"
)

func TestExtractTwitchAccessToken(t *testing.T) {
	tests := map[string]string{
		"access_token=73d0f8mkabpbmjp921asv2jaidwxn": "73d0f8mkabpbmjp921asv2jaidwxn",
		"":                                     "",
		"code=NApCCg..BkWtQ&state=34fFs29kd09": "",
	}

	for input, expected := range tests {
		t.Run(fmt.Sprintf("%s=>%s", input, expected), func(t *testing.T) {
			assert.Equal(t, expected, handler.ExtractTwitchAccessCode(input))
		})
	}
}

package site

import (
	"fmt"
	"html/template"
	"io"
	"net/http"

	"github.com/saxypandabear/twitchsongrequests/internal/constants"
	"github.com/saxypandabear/twitchsongrequests/internal/util"
	"github.com/saxypandabear/twitchsongrequests/pkg/db"
	"go.uber.org/zap"
)

var homePage = template.Must(template.ParseFiles("pkg/site/home.html"))

type HomePageRenderer struct {
	userStore db.UserStore
	twitch    *util.AuthConfig
	spotify   *util.AuthConfig
	siteURL   string
}

type HomePageData struct {
	UserID         string
	TwitchAuthURL  string
	SubscribeURL   string
	UnsubscribeURL string
	SpotifyAuthURL string
	PreferencesURL string
	Authenticated  bool
	Subscribed     bool
	Error          string
	BrowserSource  string
}

func NewHomePageRenderer(siteURL string, u db.UserStore, twitch, spotify *util.AuthConfig) *HomePageRenderer {
	return &HomePageRenderer{
		siteURL:   siteURL,
		userStore: u,
		twitch:    twitch,
		spotify:   spotify,
	}
}

func (h *HomePageRenderer) HomePage(w http.ResponseWriter, r *http.Request) {
	data := h.getHomePageData(r)

	if err := homePage.Execute(w, data); err != nil {
		zap.L().Error("error occurred while executing template", zap.Error(err))
	}
}

func (h *HomePageRenderer) getHomePageData(r *http.Request) *HomePageData {
	d := HomePageData{
		SubscribeURL:   fmt.Sprintf("%s/subscribe", h.siteURL),
		UnsubscribeURL: fmt.Sprintf("%s/revoke", h.siteURL),
		PreferencesURL: fmt.Sprintf("%s/preferences", h.siteURL),
		TwitchAuthURL:  util.GenerateAuthURL("id.twitch.tv", "oauth2/authorize", h.twitch),
		SpotifyAuthURL: util.GenerateAuthURL("accounts.spotify.com", "authorize", h.spotify),
	}

	id, err := util.GetUserIDFromRequest(r)
	if err != nil {
		zap.L().Error("failed to get Twitch ID", zap.Error(err))
		return &d
	}

	d.UserID = id

	user, err := h.userStore.GetUser(id)
	if err != nil || user == nil {
		zap.L().Error("failed to get user", zap.String("id", id), zap.Error(err))
		return &d
	}

	d.Authenticated = user.IsAuthenticated()
	d.Subscribed = user.Subscribed

	if d.Subscribed {
		// They're subscribed, so display the OBS source link
		c, _ := r.Cookie(constants.TwitchIDCookieKey) // At this point, the cookie is already valid
		d.BrowserSource = fmt.Sprintf("%s/queue/%s", h.siteURL, c.Value)
	}

	// check if there is an error in the request body
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		zap.L().Error("failed to read request body", zap.Error(err))
	} else {
		// TODO: I don't think this does anything
		if len(body) > 0 {
			d.Error = string(body)
			zap.L().Error("found error in request body", zap.String("error", d.Error))
		}
	}

	return &d
}

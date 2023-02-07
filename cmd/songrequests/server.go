package songrequests

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nicklaw5/helix"
	"github.com/saxypandabear/twitchsongrequests/internal/util"
	"github.com/saxypandabear/twitchsongrequests/pkg/api"
	"github.com/saxypandabear/twitchsongrequests/pkg/constants"
	"github.com/saxypandabear/twitchsongrequests/pkg/site"
	"github.com/saxypandabear/twitchsongrequests/pkg/spotify"
)

func StartServer(port int) error {
	if port < 1 {
		log.Fatalf("Invalid port: %d", port)
	}
	addr := fmt.Sprintf(":%d", port)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	homePage := site.NewHomePageRenderer()
	r.Get("/", homePage.HomePage)

	r.Get("/ping", api.PingHandler)

	redirectURL := util.GetFromEnvOrDefault(constants.SiteRedirectURL, fmt.Sprintf("http://localhost:%s", addr))

	s, err := util.GetFromEnv(constants.TwitchEventSubSecretKey)
	if err != nil {
		log.Println("failed", err)
		return err
	}

	twitchOptions, err := util.LoadTwitchClientOptions()
	if err != nil {
		log.Println("failed to load Twitch configurations ", err)
		return err
	}
	twitch, err := helix.NewClient(twitchOptions)
	if err != nil {
		log.Println("failed to create Twitch client ", err)
		return err
	}

	spotifyOptions, err := util.LoadSpotifyClientOptions()
	if err != nil {
		log.Println("failed to load Spotify configurations ", err)
		return err
	}
	p := spotify.SpotifyPlayerQueue{
		Auth: spotifyOptions,
	}
	reward := api.NewRewardHandler(s, &p)

	r.Post("/callback", reward.ChannelPointRedeem)

	twitchRedirect := api.NewTwitchAuthZHandler(redirectURL, twitch)
	spotifyRedirect := api.NewSpotifyAuthZHandler(redirectURL, spotifyOptions)
	r.Get("/oauth/twitch", twitchRedirect.SubscribeToTopic)
	r.Get("/oauth/spotify", spotifyRedirect.Authenticate)

	http.Handle("/", r)

	srv := &http.Server{
		Handler:           r,
		Addr:              addr,
		WriteTimeout:      15 * time.Second,
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 15 * time.Second,
	}

	log.Printf("Starting server on %s\n", srv.Addr)
	return srv.ListenAndServe()
}

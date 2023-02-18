package songrequests

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nicklaw5/helix"
	"github.com/saxypandabear/twitchsongrequests/internal/util"
	"github.com/saxypandabear/twitchsongrequests/pkg/api"
	"github.com/saxypandabear/twitchsongrequests/pkg/constants"
	"github.com/saxypandabear/twitchsongrequests/pkg/db"
	"github.com/saxypandabear/twitchsongrequests/pkg/site"
	"github.com/saxypandabear/twitchsongrequests/pkg/spotify"
)

func StartServer(port int) error {
	if port < 1 {
		return fmt.Errorf("invalid port %d", port)
	}
	addr := fmt.Sprintf(":%d", port)

	// connect to Postgres DB
	dbpool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println("failed to connect to Postgres database")

	}
	defer dbpool.Close()

	userStore := db.NewPostgresUserStore(dbpool)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.CleanPath)
	r.Use(httprate.LimitByIP(100, time.Minute))
	r.Use(middleware.Recoverer)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(middleware.Heartbeat("/ping"))

	r.NotFound(site.NotFound)
	r.MethodNotAllowed(site.NotAllowed)

	redirectURL := util.GetFromEnvOrDefault(constants.SiteRedirectURL, fmt.Sprintf("http://localhost:%s", addr))

	s, err := util.GetFromEnv(constants.TwitchEventSubSecretKey)
	if err != nil {
		log.Println("failed to load Twitch auth state key", err)
		return err
	}

	twitchState := util.GetFromEnvOrDefault(constants.TwitchStateKey, "foo123")
	spotifyState := util.GetFromEnvOrDefault(constants.SpotifyStateKey, "bar789")

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

	p := spotify.SpotifyPlayerQueue{}
	reward := api.NewRewardHandler(s, &p, userStore)

	r.Post("/callback", reward.ChannelPointRedeem)

	twitchRedirect := api.NewTwitchAuthZHandler(redirectURL, twitchState, twitch, userStore)
	spotifyRedirect := api.NewSpotifyAuthZHandler(redirectURL, spotifyState, spotifyOptions, userStore)
	r.Get("/oauth/twitch", twitchRedirect.SubscribeToTopic)
	r.Get("/oauth/spotify", spotifyRedirect.Authenticate)

	twitchConfig := site.AuthConfig{
		ClientID:    twitchOptions.ClientID,
		RedirectURL: twitchOptions.RedirectURI,
		State:       twitchState,
	}

	pageHandler := site.NewSiteRenderer(userStore, twitchConfig)
	r.Get("/", pageHandler.HomePage)

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

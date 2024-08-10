package songrequests

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/saxypandabear/twitchsongrequests/internal/constants"
	"github.com/saxypandabear/twitchsongrequests/internal/util"
	"github.com/saxypandabear/twitchsongrequests/pkg/api"
	"github.com/saxypandabear/twitchsongrequests/pkg/db"
	"github.com/saxypandabear/twitchsongrequests/pkg/o11y/logger"
	"github.com/saxypandabear/twitchsongrequests/pkg/site"
	"github.com/saxypandabear/twitchsongrequests/pkg/spotify"
	"go.uber.org/zap"
)

func StartServer(zaplogger *zap.Logger, port int) error {
	if port < 1 {
		return fmt.Errorf("invalid port %d", port)
	}
	addr := fmt.Sprintf(":%d", port)

	// create database first, checking flag if we do a no-op here or use a Postgres-backed
	// database.
	_, ok := os.LookupEnv(constants.SkipPostgres) // only care if the flag exists in the environment
	var userStore db.UserStore
	var preferenceStore db.PreferenceStore
	var messageCounter db.MessageCounter
	if !ok {
		// use no-op implementations
		userStore = &db.NoopUserStore{}
		preferenceStore = &db.NoopPreferenceStore{}
		messageCounter = &db.NoopMessageCounter{}
	} else {
		// connect to Postgres DB
		dbpool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
		if err != nil {
			return fmt.Errorf("failed to connect to Postgres database")
		}
		defer dbpool.Close()

		userStore = db.NewPostgresUserStore(dbpool)
		preferenceStore = db.NewPostgresPreferenceStore(dbpool)
		messageCounter = db.NewPostgresMessageCounter(dbpool)
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.CleanPath)
	r.Use(middleware.RequestLogger(&logger.ZapFormatter{L: zaplogger}))
	r.Use(httprate.LimitByIP(10000, time.Minute))
	r.Use(middleware.Recoverer)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(middleware.Heartbeat("/ping"))

	fallbackURL := util.GetFromEnvOrDefault(constants.RailwayDomain, fmt.Sprintf("http://localhost%s", addr))
	redirectURL := util.GetFromEnvOrDefault(constants.SiteRedirectURL, fallbackURL)

	s, err := util.GetFromEnv(constants.TwitchEventSubSecretKey)
	if err != nil {
		zap.L().Error("failed to load Twitch auth state key", zap.Error(err))
		return err
	}

	twitchConfig, err := util.LoadTwitchConfigs()
	if err != nil {
		zap.L().Error("failed to load Twitch configurations ", zap.Error(err))
		return err
	}

	spotifyConfig, err := util.LoadSpotifyConfigs()
	if err != nil {
		zap.L().Error("failed to load Spotify configurations ", zap.Error(err))
		return err
	}

	onboardedUsers := util.GetFromEnvOrDefault(constants.NumOnboardedUsers, "0")
	allowedUsers := util.GetFromEnvOrDefault(constants.NumAllowedUsers, "1")
	var numOnboarded uint
	var numAllowed uint = 1
	if i, err := strconv.Atoi(onboardedUsers); err == nil {
		numOnboarded = uint(i)
	}
	if i, err := strconv.Atoi(allowedUsers); err == nil {
		numAllowed = uint(i)
	}
	zap.L().Debug(fmt.Sprintf("Currently serving song requests for %d/%d users", numOnboarded, numAllowed))

	// ===== APIs =====
	p := spotify.NewSpotifyPlayerQueue()
	rhconfig := api.RewardHandlerConfig{
		Secret:    s,
		Publisher: p,
		UserStore: userStore,
		PrefStore: preferenceStore,
		MsgCount:  messageCounter,
		Twitch:    twitchConfig,
		Spotify:   spotifyConfig,
	}
	reward := api.NewRewardHandler(&rhconfig)

	r.Post("/callback", reward.ChannelPointRedeem)

	eventSub := api.NewEventSubHandler(userStore, preferenceStore, twitchConfig, redirectURL, s)
	r.Post("/subscribe", eventSub.SubscribeToTopic)

	twitchRedirect := api.NewTwitchAuthZHandler(redirectURL, twitchConfig, userStore, preferenceStore)
	spotifyRedirect := api.NewSpotifyAuthZHandler(redirectURL, spotifyConfig, userStore)
	r.Get("/oauth/twitch", twitchRedirect.Authorize)
	r.Get("/oauth/spotify", spotifyRedirect.Authorize)

	userHandler := api.NewUserHandler(userStore, preferenceStore, redirectURL, twitchConfig, spotifyConfig)
	r.Post("/revoke", userHandler.RevokeUserAccesses) // this is a POST because forms don't support DELETE

	preferenceHandler := api.NewPreferenceHandler(preferenceStore, redirectURL)
	r.Post("/preference", preferenceHandler.SavePreferences) // this is a POST because forms don't support DELETE

	statsHandler := api.NewStatsHandler(messageCounter, numOnboarded, numAllowed)
	r.Get("/stats/total", statsHandler.TotalMessages)
	r.Get("/stats/running", statsHandler.RunningCount)
	r.Get("/stats/onboarded", statsHandler.Onboarded)

	queueHandler := site.NewQueuePageRenderer(redirectURL, userStore, spotifyConfig)
	r.Get("/queue/{id}", queueHandler.GetUserQueue)

	// ===== Website Pages =====

	home := site.NewHomePageRenderer(redirectURL, userStore, twitchConfig, spotifyConfig)
	preferences := site.NewPreferencesRenderer(preferenceStore, redirectURL)
	r.Get("/", home.HomePage)
	r.Get("/preferences", preferences.PreferencesPage)

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

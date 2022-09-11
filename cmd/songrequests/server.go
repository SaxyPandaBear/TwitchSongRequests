package songrequests

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/nicklaw5/helix"
	"github.com/saxypandabear/twitchsongrequests/pkg/constants"
	"github.com/saxypandabear/twitchsongrequests/pkg/handler"
	"github.com/saxypandabear/twitchsongrequests/pkg/spotify"
)

func StartServer(port int) error {
	if port < 1 {
		log.Fatalf("Invalid port: %d", port)
	}
	addr := fmt.Sprintf(":%d", port)

	r := mux.NewRouter()
	r.HandleFunc("/", handler.PingHandler).Methods("GET")

	twitchOptions, err := loadTwitchClientOptions()
	if err != nil {
		log.Println("failed to load Twitch configurations ", err)
		return err
	}
	twitch, err := helix.NewClient(twitchOptions)
	if err != nil {
		log.Println("failed to create Twitch client ", err)
		return err
	}
	eventSub := handler.NewEventSubHandler(twitch)

	s, err := getEventSubSecret()
	if err != nil {
		log.Println("failed", err)
		return err
	}
	p := spotify.SpotifyPlayeQueue{}
	reward := handler.NewRewardHandler(s, p)

	r.HandleFunc("/subscribe", eventSub.SubscribeToTopic).Methods("POST")
	r.HandleFunc("/callback", reward.ChannelPointRedeem)

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

// loadTwitchClientOptions reads from environment variables in order to populate
// configuration options for the Twitch client.
func loadTwitchClientOptions() (*helix.Options, error) {
	clientID, ok := os.LookupEnv(constants.TwitchClientIDKey)
	if !ok {
		return nil, fmt.Errorf("%s is not defined in the environment", constants.TwitchClientIDKey)
	} else if len(clientID) < 1 {
		return nil, fmt.Errorf("%s is empty", constants.TwitchClientIDKey)
	}

	opt := helix.Options{
		ClientID: clientID,
	}

	// handle case where we are running locally with the Twitch CLI mock server
	url, ok := os.LookupEnv(constants.MockServerURLKey)
	if ok && len(url) > 0 {
		log.Printf("Using mocked Twitch API hosted at %s\n", url)
		opt.APIBaseURL = url
	}

	return &opt, nil
}

func getEventSubSecret() (string, error) {
	s, ok := os.LookupEnv(constants.TwitchEventSubSecret)
	if !ok {
		return "", fmt.Errorf("%s is not defined in the environment", constants.TwitchEventSubSecret)
	} else if len(s) < 1 {
		return "", fmt.Errorf("%s is empty", constants.TwitchEventSubSecret)
	}

	return s, nil
}

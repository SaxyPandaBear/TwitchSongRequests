package handler

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/nicklaw5/helix"
)

const (
	topicType    = "channel.channel_points_custom_reward_redemption.add"
	topicVersion = "1"
	subMethod    = "webhook"
)

type SubscribeRequest struct {
	UserID string `json:"user_id"`
}

type EventSubHandler struct {
	client *helix.Client
}

func NewEventSubHandler(c *helix.Client) EventSubHandler {
	return EventSubHandler{
		client: c,
	}
}

// SubscribeToTopic
func (e *EventSubHandler) SubscribeToTopic(w http.ResponseWriter, r *http.Request) {
	var req SubscribeRequest
	b, err := io.ReadAll(r.Body)

	if err != nil || len(b) < 1 {
		log.Println("failed to read request body ", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(b, &req)
	if err != nil {
		log.Println("failed to unmarshal request ", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	createSub := helix.EventSubSubscription{
		Type:    topicType,
		Version: topicVersion,
		Condition: helix.EventSubCondition{
			BroadcasterUserID: req.UserID,
		},
		Transport: helix.EventSubTransport{
			Method: subMethod,
		},
	}
	_, err = e.client.CreateEventSubSubscription(&createSub)
	if err != nil {
		log.Println("failed to create EventSub subscription ", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Println("successfully subscribed to Channel Point topic for user ")
	w.WriteHeader(http.StatusCreated)
}

/**
 * Simple test suite to be run with Node.
 * Usage (from root directory): node ./js/tests.js
 */

console.log("Twitch Tests");
const testEvent = {
    "type": "reward-redeemed",
    "data": {
      "timestamp": "2019-11-12T01:29:34.98329743Z",
      "redemption": {
        "id": "9203c6f0-51b6-4d1d-a9ae-8eafdb0d6d47",
        "user": {
          "id": "30515034",
          "login": "davethecust",
          "display_name": "davethecust"
        },
        "channel_id": "30515034",
        "redeemed_at": "2019-12-11T18:52:53.128421623Z",
        "reward": {
          "id": "6ef17bb2-e5ae-432e-8b3f-5ac4dd774668",
          "channel_id": "30515034",
          "title": "hit a gleesh walk on stream",
          "prompt": "cleanside's finest \n",
          "cost": 10,
          "is_user_input_required": true,
          "is_sub_only": false,
          "image": {
            "url_1x": "https://static-cdn.jtvnw.net/custom-reward-images/30515034/6ef17bb2-e5ae-432e-8b3f-5ac4dd774668/7bcd9ca8-da17-42c9-800a-2f08832e5d4b/custom-1.png",
            "url_2x": "https://static-cdn.jtvnw.net/custom-reward-images/30515034/6ef17bb2-e5ae-432e-8b3f-5ac4dd774668/7bcd9ca8-da17-42c9-800a-2f08832e5d4b/custom-2.png",
            "url_4x": "https://static-cdn.jtvnw.net/custom-reward-images/30515034/6ef17bb2-e5ae-432e-8b3f-5ac4dd774668/7bcd9ca8-da17-42c9-800a-2f08832e5d4b/custom-4.png"
          },
          "default_image": {
            "url_1x": "https://static-cdn.jtvnw.net/custom-reward-images/default-1.png",
            "url_2x": "https://static-cdn.jtvnw.net/custom-reward-images/default-2.png",
            "url_4x": "https://static-cdn.jtvnw.net/custom-reward-images/default-4.png"
          },
          "background_color": "#00C7AC",
          "is_enabled": true,
          "is_paused": false,
          "is_in_stock": true,
          "max_per_stream": { "is_enabled": false, "max_per_stream": 0 },
          "should_redemptions_skip_request_queue": true
        },
        "user_input": "yeooo",
        "status": "FULFILLED"
        }
      }
    }

console.log("Spotify Tests");
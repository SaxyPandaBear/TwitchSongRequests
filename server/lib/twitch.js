const WebSocket = require('ws');
const AWS = require('aws-sdk');
const { awsConfig } = require('../session');

var sqsClient = new AWS.SQS(awsConfig);

// We require the Queue URL in order to know what queue to send the message to.
// this should be defined in environment variables.
// Note: This env var should be injected by the cloudformation template when it
//       stands up the whole infrastructure.
const queueUrl = process.env['QUEUE_URL'];

const spotifyTrackPattern = /(spotify:track:(\w||\d)+)/y;

// we have to ensure that there is at least some indication that the channel point
// redemption is related to a song request. we are declaring that a song request
// that we act on is something that has "song request" in the title
function isSongRequest(rewardTitle) {
    return rewardTitle.toLowerCase().includes('song request');
}

// spotify:track:5Cjkfft7iRWJp4elZXgjkc
function matchesSpotifyUri(uri) {
    const matches = uri.match(spotifyTrackPattern);
    // if it fails to match at all, the result of matches will be null
    if (matches) {
        // if something matches, the regex we have still allows for extraneous input
        // after the match, so just do a sanity check on the raw input.
        // Because of how the regex is structured, we can pattern match on whitespace.
        // if we split on whitespace, we should only get one result back.
        return uri.split(/(\s+)/).length === 1;
    } else {
        return false;
    }
}

function openSocketConnectionWithChannelId(channelId, oauthToken, callBack) {
    ws = new WebSocket('wss://pubsub-edge.twitch.tv');

    function heartbeat() {
        message = {
            type: 'PING',
        };
        console.log('SENT: ' + JSON.stringify(message));
        ws.send(JSON.stringify(message));
    }
    var heartbeatInterval = 1000 * 60; //ms between PING's
    var reconnectInterval = 1000 * 3; //ms to wait before reconnect
    var heartbeatHandle;
    ws.onopen = function (event) {
        console.info('Socket Opened');
        heartbeat();
        heartbeatHandle = setInterval(heartbeat, heartbeatInterval);

        // When the connection is opened, fire a LISTEN event to the WSS
        let listenEvent = {
            type: 'LISTEN',
            nonce: 'abc123',
            data: {
                topics: [`channel-points-channel-v1.${channelId}`],
                auth_token: oauthToken,
            },
        };
        ws.send(JSON.stringify(listenEvent));

        //DONE to signal succesfull connection -- another way to handle this is to direclty inject the DAO here but that's too tightly coupled
        callBack();
    };

    ws.onerror = function (error) {
        console.log('ERR: ' + JSON.stringify(error));
    };

    ws.onmessage = function (event) {
        let message = JSON.parse(event.data);
        if (message.type === 'RECONNECT') {
            console.info('Reconnecting...');
            setTimeout(
                openSocketConnectionWithChannelId(channelId),
                reconnectInterval
            );
        } else if (message.type === 'MESSAGE') {
            // The data is deeply nested. Bear with me.
            const innerMessage = JSON.parse(message.data.message);
            const redemption = innerMessage.data.redemption;
            const rewardTitle = redemption.reward.title;
            const spotifyUri = redemption.user_input;
            if (isSongRequest(rewardTitle) && matchesSpotifyUri(spotifyUri)) {
                // if the input is validated, send a message to SQS so that the
                // lambda can pick up work on this.
                const params = {
                    MessageBody: spotifyUri,
                    QueueUrl: queueUrl,
                    MessageAttributes: {
                        channelId: {
                            DataType: 'String',
                            StringValue: channelId,
                        },
                    },
                };
                // TODO: clean this up
                sqsClient.sendMessage(params, function (err, data) {
                    if (err) console.log(err, err.stack);
                    // an error occurred
                    else console.log(data); // successful response
                });
            }
        }
    };

    ws.onclose = function () {
        console.info('Socket Closed');
        clearInterval(heartbeatHandle);
        console.info('Reconnecting...');
        setTimeout(
            openSocketConnectionWithChannelId(channelId),
            reconnectInterval
        );
    };
}
module.exports = openSocketConnectionWithChannelId;

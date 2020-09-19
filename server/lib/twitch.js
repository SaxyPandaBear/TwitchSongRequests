const fetch = require('node-fetch');
const WebSocket = require('ws');

//ANDREW PLEASE
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
        // TODO: add Spotify integration here
        message = JSON.parse(event.data);
        console.log({ message });
        if (message.type == 'RECONNECT') {
            console.info('Reconnecting...');
            setTimeout(
                openSocketConnectionWithChannelId(channelId),
                reconnectInterval
            );
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

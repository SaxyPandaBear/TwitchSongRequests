const {
    fetchSpotifyToken,
    fetchTwitchToken,
    fetchTwitchChannel,
} = require('../lib/auth');

const {
    spotifyClientId,
    spotifyClientSecret,
    twitchClientSecret,
    twitchClientId,
} = require('../config/credentials.json');

const SessionAuthController = {
    postClientTwitchAccessCode(req, res) {
        const {
            body: { accessKey },
        } = req;

        fetchTwitchToken(twitchClientId, twitchClientSecret, accessKey)
            .then((twitchResponse) => {
                const {
                    access_token,
                    refresh_token,
                    token_type,
                } = twitchResponse;

                req.session.accessKeys.twitchToken = {
                    access_token,
                    refresh_token,
                    token_type,
                };

                return fetchTwitchChannel(access_token, twitchClientId);
            })
            .then((channelResponse) => {
                const { _id: channelId } = channelResponse;

                req.session.accessKeys.twitchToken.channelId = channelId;
                res.status(200).json({ success: true });
            })
            .catch((err) => {
                console.log({ err });
            });
    },
    postClientSpotifyAccessCode(req, res) {
        const {
            body: { accessKey },
        } = req;

        fetchSpotifyToken(spotifyClientId, spotifyClientSecret, accessKey)
            .then((spotifyResponse) => {
                const {
                    access_token,
                    refresh_token,
                    token_type,
                    scope,
                } = spotifyResponse;
                req.session.accessKeys.spotifyToken = {
                    access_token,
                    refresh_token,
                    scope,
                    token_type,
                };
                res.status(200).json({ success: true });
            })
            .catch((err) =>
                res
                    .status(500)
                    .json({ error: 'Something went wrong with spotify oauth' })
            );
    },
    getClientAuthStatus(req, res) {
        const { twitchToken, spotifyToken } = req.session.accessKeys;
        res.json({
            twitchToken: !!twitchToken,
            spotifyToken: !!spotifyToken,
        });
    },
};

module.exports = SessionAuthController;

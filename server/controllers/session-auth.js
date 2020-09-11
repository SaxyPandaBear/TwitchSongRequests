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
    postClientTwitchAccessCode(req, res, next) {
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

                const twitchTokenConfiguration = {
                    access_token,
                    refresh_token,
                    token_type,
                };
                req.twitchTokenConfiguration = twitchTokenConfiguration;

                return fetchTwitchChannel(access_token, twitchClientId);
            })
            .then((channelResponse) => {
                const { _id: channelId } = channelResponse;
                req.channelId = channelId;
                next();
            })
            .catch((err) => {
                console.log({ err });
            });
    },
    postClientSpotifyAccessCode(req, res, next) {
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
                const spotifyTokenConfiguration = {
                    access_token,
                    refresh_token,
                    token_type,
                    scope,
                };
                req.spotifyTokenConfiguration = spotifyTokenConfiguration;
                next();
            })
            .catch((err) =>
                res
                    .status(500)
                    .json({ error: 'Something went wrong with spotify oauth' })
            );
    },
    getClientAuthStatus(req, res) {
        if (req.session) {
            const { twitchToken, spotifyToken } = req.session.accessKeys;
            res.json({
                twitchToken: !!twitchToken,
                spotifyToken: !!spotifyToken,
            });
        } else {
            res.json({
                twitchToken: false,
                spotifyToken: false,
            });
        }
    },
};

module.exports = SessionAuthController;

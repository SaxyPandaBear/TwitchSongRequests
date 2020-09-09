const {
  fetchSpotifyToken,
  fetchTwitchToken,
  fetchTwitchChannel,
} = require("../auth");

const {
  spotifyClientId,
  spotifyClientSecret,
  twitchClientSecret,
  twitchClientId,
} = require("../config/credentials.json");

const SessionAuthController = {
  postClientTwitchAccessCode(req, res) {
    const {
      body: { accessKey },
    } = req;

    fetchTwitchToken(twitchClientId, twitchClientSecret, accessKey)
      .then((twitchResponse) => {
        const {
          access_token,
          expires_in,
          refresh_token,
          token_type,
        } = twitchResponse;

        req.session.accessKeys.twitchToken = {
          access_token,
          expires_in,
          refresh_token,
          token_type,
          expirationDate: new Date(new Date().getTime() + expires_in * 1000),
        };

        return fetchTwitchChannel(access_token, twitchClientId);
      })
      .then((channelResponse) => {
        console.log({ channelResponse });
        const { _id: channelId } = channelResponse;
        req.session.accessKeys.twitchToken.channelId = channelId;
        res.status(200).json({ success: true });
      })
      .catch((err) => {
        console.log({ err });
      });

    // TODO: Invoke auth endpoint and generate proper token
    // req.session.twitchChannelId = channelId;
  },
  postClientSpotifyAccessCode(req, res) {
    const {
      body: { accessKey },
    } = req;

    fetchSpotifyToken(spotifyClientId, spotifyClientSecret, accessKey)
      .then((spotifyResponse) => {
        const {
          access_token,
          expires_in,
          refresh_token,
          token_type,
          scope,
        } = spotifyResponse;
        req.session.accessKeys.spotifyToken = {
          access_token,
          expires_in,
          refresh_token,
          scope,
          token_type,
          expirationDate: new Date(new Date().getTime() + expires_in * 1000),
        };
        console.log({
          spotify: req.session.accessKeys.spotifyToken,
          twitch: req.session.accessKeys.twitchToken,
        });
        res.status(200).json({ success: true });
      })
      .catch((err) =>
        res
          .status(500)
          .json({ error: "Something went wrong with spotify oauth" })
      );
  },
  getClientAuthStatus(req, res) {
    const { twitchToken, spotifyToken } = req.session.accessKeys;
    console.log(req.session.accessKeys);
    res.json({
      twitchToken: !!twitchToken,
      spotifyToken: !!spotifyToken,
    });
  },
};

module.exports = SessionAuthController;

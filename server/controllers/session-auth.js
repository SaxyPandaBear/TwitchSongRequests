const SessionAuthController = {
  postClientTwitchAccessCode(req, res) {
    const {
      body: { accessKey },
    } = req;
    req.session.accessKeys.twitchAccessKey = accessKey;
    res.status(200).json({ success: true });
  },
  postClientSpotifyAccessCode(req, res) {
    const {
      body: { accessKey },
    } = req;
    req.session.accessKeys.spotifyAccessKey = accessKey;
    res.status(200).json({ success: true });
  },
  getClientAuthStatus(req, res) {
    const { twitchAccessKey, spotifyAccessKey } = req.session.accessKeys;
    res.json({
      twitchAccessKey: !!twitchAccessKey,
      spotifyAccessKey: !!spotifyAccessKey,
    });
  },
};
// post("/oauth/twitch", function (req, res, next) {
//   const {
//     body: { accessKey },
//   } = req;
//   req.session.accessKeys.twitchAccessKey = accessKey;
//   console.log({ session: req.session });

//   res.status(200).json({ success: true });
// });
// app.post("/oauth/spotify", function (req, res, next) {
//   const {
//     body: { accessKey },
//   } = req;
//   req.session.accessKeys.spotifyAccessKey = accessKey;
//   console.log({ session: req.session });
//   res.status(200).json({ success: true });
// });

// app.get("/oauth/access-keys", (req, res) => {
//   const { twitchAccessKey, spotifyAccessKey } = req.session.accessKeys;
//   res.json({
//     twitchAccessKey: !!twitchAccessKey,
//     spotifyAccessKey: !!spotifyAccessKey,
//   });
// });
module.exports = SessionAuthController;

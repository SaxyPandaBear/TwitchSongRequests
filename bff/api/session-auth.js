const express = require('express');
const {
    intializeSesionStore,
    checkForExistingSessionAndAssignAccessKeys,
    assignTwitchTokenToSession,
    assignSpotifyTokenToSession,
} = require('../session');
const sessionAuthController = require('../controllers/session-auth');
const { updateConnectionStatusByChannelId } = require('../lib/dynamoDao');
const router = express.Router();
router.post(
    '/twitch',
    sessionAuthController.postClientTwitchAccessCode,
    intializeSesionStore(),
    checkForExistingSessionAndAssignAccessKeys,
    assignTwitchTokenToSession,
    (req, res, next) => {
        const { channelId } = req;
        updateConnectionStatusByChannelId(channelId, 'starting')
            .then((status) => {
                console.log({ status });
                next();
            })
            .catch(console.error);
    },
    (req, res, next) => {
        res.status(200).json({ success: true });
        next();
    },
    sessionAuthController.connectToTwitchChat
);
router.post(
    '/spotify',
    sessionAuthController.postClientSpotifyAccessCode,
    assignSpotifyTokenToSession,
    (req, res) => {
        res.status(200).json({ success: true });
    }
);
router.get('/access-keys', sessionAuthController.getClientAuthStatus);
module.exports = router;

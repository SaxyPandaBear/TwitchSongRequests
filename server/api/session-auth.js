const express = require('express');
const {
    intializeSesionStore,
    checkForExistingSessionAndAssignAccessKeys,
    assignTwitchTokenToSession,
    assignSpotifyTokenToSession,
} = require('../session');
const router = express.Router();
const sessionAuthController = require('../controllers/session-auth');
router.post(
    '/twitch',
    sessionAuthController.postClientTwitchAccessCode,
    intializeSesionStore(),
    checkForExistingSessionAndAssignAccessKeys,
    assignTwitchTokenToSession,
    (req, res) => {
        res.status(200).json({ success: true });
    }
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

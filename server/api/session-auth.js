const express = require('express');
const router = express.Router();
const sessionAuthController = require('../controllers/session-auth');
router.post('/twitch', sessionAuthController.postClientTwitchAccessCode);
router.post('/spotify', sessionAuthController.postClientSpotifyAccessCode);
router.get('/access-keys', sessionAuthController.getClientAuthStatus);
module.exports = router;

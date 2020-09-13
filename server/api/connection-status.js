const express = require('express');
const ConnectionStatusController = require('../controllers/connection-status');
const { route } = require('./session-auth');
const router = express.Router();
router.get('/:channelId', ConnectionStatusController.getConnectionStatus);

module.exports = router;

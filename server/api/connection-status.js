const express = require('express');
const ConnectionStatusController = require('../controllers/connection-status');
const router = express.Router();
router.get('/:channelId', ConnectionStatusController.getConnectionStatus);
router.post('/:channelId', ConnectionStatusController.updateConnectionStatus);

module.exports = router;

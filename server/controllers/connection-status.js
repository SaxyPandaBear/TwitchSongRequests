const {
    queryDynamoByChannel,
    updateConnectionStatusByChannelId,
} = require('../lib/dynamoDao');

const ConnectionStatusController = {
    getConnectionStatus(req, res) {
        const channelId = req.params.channelId;
        console.log({ channelId });

        queryDynamoByChannel(channelId)
            .then((item) => {
                if (!item) {
                    res.status(404).json({
                        message: 'No items found for provided channelId',
                    });
                } else {
                    res.json({ item });
                }
            })
            .catch((err) => {
                console.log(err);
                res.status(500).json({ err });
            });
    },
    updateConnectionStatus(req, res) {
        const channelId = req.params.channelId;
        const { connectionStatus } = req.body;
        if (this.validConnectionStatus(connectionStatus)) {
            updateConnectionStatusByChannelId(channelId, connectionStatus)
                .then((item) => {
                    res.json({ item });
                })
                .catch((err) => {
                    res.status(500).json({ err });
                });
        } else {
            return res
                .status(400)
                .json({ err: 'Provide a valid connection status' });
        }
    },
    isValidConnectionStatus: function (connectionStatus) {
        return Object.values(this.validConnectionStatus).includes(
            connectionStatus
        );
    },
    validConnectionStatus: {
        ACTIVE: 'active',
        INACTIVE: 'inactive',
        STARTING: 'starting',
    },
};

module.exports = ConnectionStatusController;

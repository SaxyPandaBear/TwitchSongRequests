const {
    queryDynamoByChannel,
    updateConnectionStatus,
} = require('../lib/dynamoDao');

// TO BE USED LATER AS SOME KIND OF TYPESAFE ENUM VALIDATION
const CONNECTION_STATUS = {
    ACTIVE: 'active',
    INACTIVE: 'inactive',
    STARTING: 'starting',
};
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
    updateConnectionStatus(channelId, connectionStatus) {
        updateConnectionStatus(channelId, connectionStatus)
            .then((item) => {
                res.json({ item });
            })
            .catch((err) => {
                res.status(500).json({ err });
            });
    },
};

//ConnectionStatusController.getConnectionStatus('577228983');
module.exports = ConnectionStatusController;

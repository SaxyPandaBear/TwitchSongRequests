Connector Backend
=================

This module is what drives a good majority of the flow for the overall service.
It accepts HTTP requests that either initiate a connection to the Twitch PubSub
API, or disconnect from a specific channel topic.

### What lives in this package?
The main server code, including the API that accepts requests to connect/disconnect,
the orchestration code to handle the Twitch socket connections.

### Build
`gradle assemble`

### Test
`gradle test`

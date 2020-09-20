package com.github.saxypandabear.songrequests.jetty.api

import jakarta.ws.rs.core.Response
import jakarta.ws.rs.{POST, PUT, Path}

/**
 * Class that deals with accepting incoming requests that we expect from a Lambda.
 * This exposes an API that handles two main things:
 *  1. Requests that initiate a WebSocket connection to the Twitch API
 *  2. Requests that tell us to disconnect from a particular channel
 */
@Path("/connect")
class ConnectionResource {

    @POST
    def initiateConnection(): Response = {
        Response.status(201, "Successfully initiated connection").build()
    }

    @PUT
    def disconnect(): Response = {
        Response.noContent().build()
    }
}

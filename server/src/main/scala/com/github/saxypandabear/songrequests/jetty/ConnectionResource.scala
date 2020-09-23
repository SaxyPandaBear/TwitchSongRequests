package com.github.saxypandabear.songrequests.jetty

import com.github.saxypandabear.songrequests.jetty.model.Channel
import javax.ws.rs.core.{MediaType, Response}
import javax.ws.rs._

/**
 * Class that deals with accepting incoming requests that we expect from a Lambda.
 * This exposes an API that handles two main things:
 *  1. Requests that initiate a WebSocket connection to the Twitch API
 *  2. Requests that tell us to disconnect from a particular channel
 */
@Path("/api")
class ConnectionResource {

    @GET
    @Path("/ping")
    @Produces(Array(MediaType.APPLICATION_JSON))
    def ping(): String = {
        "pong"
    }

    @POST
    @Path("/connect")
    @Consumes(Array(MediaType.APPLICATION_JSON))
    @Produces(Array(MediaType.TEXT_PLAIN))
    def initiateConnection(request: Channel): Response = {
        Response
            .status(201)
            .entity(s"Initiated connection to channel ${request.channelId}")
            .build()
    }

    @PUT
    @Path("/disconnect/{channel}")
    @Consumes(Array(MediaType.APPLICATION_JSON))
    def disconnect(@PathParam("channel") channel: String): Response = {
        Response.noContent().build()
    }
}

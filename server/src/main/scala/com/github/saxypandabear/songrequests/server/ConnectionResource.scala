package com.github.saxypandabear.songrequests.server

import com.github.saxypandabear.songrequests.server.model.Channel
import com.github.saxypandabear.songrequests.websocket.TwitchSocketFactory
import com.github.saxypandabear.songrequests.websocket.orchestrator.ConnectionOrchestrator
import javax.inject.Inject
import javax.ws.rs._
import javax.ws.rs.core.{MediaType, Response}

/**
 * Class that deals with accepting incoming requests that we expect from a Lambda.
 * This exposes an API that handles two main things:
 *  1. Requests that initiate a WebSocket connection to the Twitch API
 *  2. Requests that tell us to disconnect from a particular channel
 */
@Path("/api")
class ConnectionResource {

  @Inject
  var orchestrator: ConnectionOrchestrator = _

  @Inject
  var twitchSocketFactory: TwitchSocketFactory = _

  @GET
  @Path("/ping")
  @Produces(Array(MediaType.APPLICATION_JSON))
  def ping(): String =
    "pong"

  @POST
  @Path("/connect")
  @Consumes(Array(MediaType.APPLICATION_JSON))
  @Produces(Array(MediaType.TEXT_PLAIN))
  def initiateConnection(request: Channel): Response =
    if (orchestrator.connect(request.channelId, twitchSocketFactory.create)) {
      Response
        .status(Response.Status.CREATED)
        .entity(s"Initiated connection to channel ${request.channelId}")
        .build()
    } else {
      Response
        .status(Response.Status.SERVICE_UNAVAILABLE)
        .entity(s"Could not connect channel ${request.channelId}")
        .build()
    }

  @PUT
  @Path("/disconnect/{channel}")
  @Consumes(Array(MediaType.APPLICATION_JSON))
  def disconnect(@PathParam("channel") channel: String): Response = {
    var success = true
    try orchestrator.disconnect(channel)
    catch {
      case _: Exception => success = false
    }
    if (success) {
      Response.noContent().build()
    } else {
      Response
        .status(Response.Status.INTERNAL_SERVER_ERROR)
        .entity(s"Did not successfully disconnect $channel")
        .build()
    }
  }
}

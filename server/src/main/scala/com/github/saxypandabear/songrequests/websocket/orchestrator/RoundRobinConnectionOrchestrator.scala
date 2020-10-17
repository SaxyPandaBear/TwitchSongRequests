package com.github.saxypandabear.songrequests.websocket.orchestrator

// TODO: implement the stuff in here
class RoundRobinConnectionOrchestrator extends ConnectionOrchestrator {

  /**
   * Initiate a connection to Twitch with an internal WebSocket connection
   * @param channelId Twitch Channel ID to listen on
   */
  override def connect(channelId: String): Unit = ???

  /**
   * Stop listening to a connection to Twitch
   * @param channelId Twitch Channel ID to stop listening on
   */
  override def disconnect(channelId: String): Unit = ???

  /**
   * Reconnect/bounce the WebSocket client to force it to reconnect, because
   * a connector received a reconnect event from the server
   * @param channelId
   */
  override def reconnect(channelId: String): Unit = ???

  /**
   * When an orchestrator is at capacity, the system should know to start
   * an auto-scaling event
   * @return true if the orchestrator is at capacity, false otherwise
   */
  override def atCapacity: Boolean = ???
}

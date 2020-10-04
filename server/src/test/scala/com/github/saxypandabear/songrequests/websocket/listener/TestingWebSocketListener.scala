package com.github.saxypandabear.songrequests.websocket.listener

import org.eclipse.jetty.websocket.api.Session

import scala.collection.mutable

/**
 * Captures and stores state based on the events that the listener receives,
 * in internal data structures so that we can inspect what events are triggered from
 * the WebSocket handler.
 *
 * The maps will associate a channel ID with the relevant info for the event.
 * For example, the map that captures close events need to capture the status code and the reason.
 * The connect event only needs to capture the channel ID, so there's not really much to map it to.
 * For simplicity, we can just store those events in a List.
 */
class TestingWebSocketListener extends WebSocketListener {

    private val lockObject = new Object()
    val connectEvents = new mutable.ArrayBuffer[String]()
    val closeEvents = new mutable.HashMap[String, (Int, String)]()
    val messageEvents = new mutable.HashMap[String, String]()
    val errorEvents = new mutable.HashMap[String, Throwable]()

    override def onConnectEvent(channelId: String, session: Session): Unit = {
        lockObject.synchronized {
            connectEvents += channelId
        }
    }

    override def onCloseEvent(channelId: String, session: Session, statusCode: Int, reason: String): Unit = {
        lockObject.synchronized {
            closeEvents.put(channelId, (statusCode, reason))
        }
    }

    override def onMessageEvent(channelId: String, session: Session, message: String): Unit = {
        lockObject.synchronized {
            messageEvents.put(channelId, message)
        }
    }

    override def onErrorEvent(channelId: String, session: Session, error: Throwable): Unit = {
        lockObject.synchronized {
            errorEvents.put(channelId, error)
        }
    }

    /**
     * Clear out all of the events that are currently stored in this listener
     */
    def flush(): Unit = {
        lockObject.synchronized {
            connectEvents.clear()
            closeEvents.clear()
            messageEvents.clear()
            errorEvents.clear()
        }
    }
}

package com.github.saxypandabear.songrequests.websocket.listener

import scala.collection.concurrent.TrieMap

/**
 * Captures and stores state based on the events that the listener receives
 */
class InspectWebSocketListener {
    val connectEvents: TrieMap[String]
}

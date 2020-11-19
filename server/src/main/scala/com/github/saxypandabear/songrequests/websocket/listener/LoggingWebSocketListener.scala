package com.github.saxypandabear.songrequests.websocket.listener
import com.typesafe.scalalogging.LazyLogging
import org.eclipse.jetty.websocket.api.Session

/**
 * Just do some logging when events come in
 */
class LoggingWebSocketListener extends WebSocketListener with LazyLogging {
  override def onConnectEvent(channelId: String, session: Session): Unit =
    logger.info(
        s"Connect event happened for $channelId on ${session.getRemoteAddress.getHostName}"
    )

  override def onCloseEvent(
      channelId: String,
      session: Session,
      statusCode: Int,
      reason: String
  ): Unit =
    logger.info(
        s"Close event happened for $channelId on ${session.getRemoteAddress.getHostName}: Status code = $statusCode, Reason = $reason"
    )

  override def onMessageEvent(
      channelId: String,
      session: Session,
      message: String
  ): Unit =
    logger.info(
        s"Message event happened for $channelId on ${session.getRemoteAddress.getHostName}: Message = $message"
    )

  override def onErrorEvent(
      channelId: String,
      session: Session,
      error: Throwable
  ): Unit =
    logger.info(
        s"Error event happened for $channelId on ${session.getRemoteAddress.getHostName}: Cause: ${error.getMessage}"
    )
}

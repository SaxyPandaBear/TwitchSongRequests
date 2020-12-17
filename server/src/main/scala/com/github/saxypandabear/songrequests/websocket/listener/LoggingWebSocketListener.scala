package com.github.saxypandabear.songrequests.websocket.listener
import com.github.saxypandabear.songrequests.types.Types.ChannelId
import com.typesafe.scalalogging.LazyLogging
import org.eclipse.jetty.websocket.api.Session

/**
 * Just do some logging when events come in
 */
class LoggingWebSocketListener extends WebSocketListener with LazyLogging {
  override def onConnectEvent(channelId: ChannelId, session: Session): Unit =
    logger.info(
        s"Connect event happened for $channelId on ${session.getRemoteAddress.getHostName}"
    )

  override def onCloseEvent(
      channelId: ChannelId,
      session: Session,
      statusCode: Int,
      reason: String
  ): Unit =
    logger.info(
        s"Close event happened for $channelId on ${session.getRemoteAddress.getHostName}: Status code = $statusCode, Reason = $reason"
    )

  override def onMessageEvent(
      channelId: ChannelId,
      session: Session,
      message: String
  ): Unit =
    logger.info(
        s"Message event happened for $channelId on ${session.getRemoteAddress.getHostName}: Message = $message"
    )

  override def onErrorEvent(
      channelId: ChannelId,
      session: Session,
      error: Throwable
  ): Unit =
    if (session == null) {
      logger.warn("Session is null for {}", channelId)
      logger.error("Error event happened for {}", channelId, error)
    } else if (session.getRemoteAddress == null) {
      logger.warn("Session remote address is null {}", channelId)
      logger.error("Error event happened for {}", channelId, error)
    } else {

      logger.info(
          "Error event happened for {} on {}",
          channelId,
          session.getRemoteAddress.getHostName,
          error
      )
    }
}

package com.github.saxypandabear.songrequests.ddb
import com.amazonaws.services.dynamodbv2.model.ResourceNotFoundException
import com.github.saxypandabear.songrequests.ddb.model.Connection

import scala.collection.concurrent.TrieMap

class InMemoryConnectionDataStore extends ConnectionDataStore {
  private val idsToConnections = new TrieMap[String, Connection]()

  /**
   * Fetch the connection details for a user by their channel ID, which is
   * the primary key (or hash key for DynamoDB)
   * This should always get the most up-to-date value of the data (a consistent read)
   * @param channelId the Twitch channel ID
   * @return a POJO that represents the connection document
   * @throws ResourceNotFoundException when the channelId does not exist in the data store
   */
  override def getConnectionDetailsById(channelId: String): Connection =
    idsToConnections.synchronized {
      idsToConnections.getOrElse(
          channelId,
          throw new ResourceNotFoundException(s"$channelId does not exist")
      )
    }

  /**
   * Update the `connectionStatus` attribute of the DynamoDB record with the
   * given input value. This should not throw an exception if the channelId
   * * does not exist.
   * @param channelId hash key of record to update
   * @param newStatus status value to persist
   */
  override def updateConnectionStatus(
      channelId: String,
      newStatus: String
  ): Unit =
    idsToConnections.synchronized {
      val connectionOpt = idsToConnections.get(channelId)
      if (connectionOpt.isDefined) {
        val updated = connectionOpt.get.copy(connectionStatus = newStatus)
        idsToConnections.put(channelId, updated)
      }
    }

  /**
   * Insert the input access token in the session object for the Twitch access
   * token and persist the object.
   * @param channelId   hash key of record to update
   * @param accessToken refreshed access key from Twitch
   * @throws ResourceNotFoundException when the channelId does not exist in the data store
   */
  override def updateTwitchOAuthToken(
      channelId: String,
      accessToken: String
  ): Unit =
    idsToConnections.synchronized {
      idsToConnections
        .get(channelId)
        .foreach(_.updateTwitchAccessToken(accessToken))
    }

  /**
   * Checks if a record with the given ID exists.
   * @param channelId the Twitch channel ID
   * @return true if the channel ID exists, false otherwise
   */
  override def hasConnectionDetails(channelId: String): Boolean =
    idsToConnections.keys.exists(_ == channelId)

  // don't need to do anything
  override def stop(): Unit = {}

  /**
   * Write a connection object to the data store. This is not a requirement for the main functionality,
   * and is only necessary for tests, which is why this method does not exist in the parent trait.
   * @param channelId  the Twitch channel ID
   * @param connection connection object
   */
  def putConnectionDetails(channelId: String, connection: Connection): Unit =
    idsToConnections.synchronized {
      idsToConnections.putIfAbsent(channelId, connection)
    }

  def clear(): Unit = idsToConnections.clear()
}

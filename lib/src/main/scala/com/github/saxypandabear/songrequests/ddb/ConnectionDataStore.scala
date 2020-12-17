package com.github.saxypandabear.songrequests.ddb

import com.amazonaws.services.dynamodbv2.model.ResourceNotFoundException
import com.github.saxypandabear.songrequests.ddb.model.Connection
import com.github.saxypandabear.songrequests.types.Types.ChannelId

/**
 * A data store that manages details about a user's connection to our services
 * This is basically an abstraction around our database usage
 */
trait ConnectionDataStore {

  /**
   * Fetch the connection details for a user by their channel ID, which is
   * the primary key (or hash key for DynamoDB).
   * This should always get the most up-to-date value of the data (a consistent read)
   * @param channelId the Twitch channel ID
   * @return a POJO that represents the connection document
   * @throws ResourceNotFoundException when the channelId does not exist in the data store
   */
  def getConnectionDetailsById(channelId: ChannelId): Connection

  /**
   * Update the `connectionStatus` attribute of the DynamoDB record with the
   * given input value. This should not throw an exception if the channelId
   * does not exist.
   * @param channelId hash key of record to update
   * @param newStatus status value to persist
   */
  def updateConnectionStatus(channelId: ChannelId, newStatus: String): Unit

  /**
   * Insert the input access token in the session object for the Twitch access
   * token and persist the object.
   * @param channelId   hash key of record to update
   * @param accessToken refreshed access key from Twitch
   * @throws ResourceNotFoundException when the channelId does not exist in the data store
   */
  def updateTwitchOAuthToken(channelId: ChannelId, accessToken: String): Unit

  /**
   * Checks if a record with the given ID exists.
   * @param channelId the Twitch channel ID
   * @return true if the channel ID exists, false otherwise
   */
  def hasConnectionDetails(channelId: ChannelId): Boolean

  def stop(): Unit
}

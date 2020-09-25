package com.github.saxypandabear.songrequests.ddb
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
     * @throws RuntimeException when the channelId does not exist in the data store
     */
    override def getConnectionDetailsById(channelId: String): Connection = {
        idsToConnections.synchronized {
            idsToConnections.getOrElse(channelId, throw new RuntimeException(s"$channelId does not exist"))
        }
    }

    /**
     * Update a record in the data store with the given hash key, and the given
     * object.
     * @param channelId  the Twitch channel ID
     * @param connection connection object
     * @throws RuntimeException when the channelId does not exist in the data store,
     *                          or the connection object is malformed
     */
    override def updateConnectionDetailsById(channelId: String, connection: Connection): Unit = {
        idsToConnections.synchronized {
            idsToConnections.put(channelId, connection)
        }
    }

    /**
     * Write a connection object to the data store. This is not a requirement for the main functionality,
     * and is only necessary for tests, which is why this method does not exist in the parent trait.
     * @param channelId  the Twitch channel ID
     * @param connection connection object
     */
    def putConnectionDetails(channelId: String, connection: Connection): Unit = {
        idsToConnections.synchronized {
            idsToConnections.put(channelId, connection)
        }
    }
}

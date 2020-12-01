package com.github.saxypandabear.songrequests.ddb
import com.amazonaws.services.dynamodbv2.AmazonDynamoDB
import com.amazonaws.services.dynamodbv2.model._
import com.github.saxypandabear.songrequests.ddb.model.Connection
import com.typesafe.scalalogging.LazyLogging

import scala.collection.JavaConverters._

class DynamoDbConnectionDataStore(dynamoDb: AmazonDynamoDB)
    extends ConnectionDataStore
    with LazyLogging {
  val TABLE_NAME = "connections"

  init()

  /**
   * Fetch the connection details for a user by their channel ID, which is
   * the primary key (or hash key for DynamoDB).
   * This should always get the most up-to-date value of the data (a consistent read)
   * @param channelId the Twitch channel ID
   * @return a POJO that represents the connection document
   * @throws ResourceNotFoundException when the channelId does not exist in the data store
   */
  override def getConnectionDetailsById(channelId: String): Connection = {
    val request = new GetItemRequest()
      .withTableName(TABLE_NAME)
      .withConsistentRead(true)
      .withKey(getHashKey(channelId).asJava)
    val item    = dynamoDb.getItem(request).getItem
    if (item == null) {
      throw new ResourceNotFoundException(s"$channelId does not exist")
    }
    Connection(item.asScala.toMap)
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
  ): Unit = {
    val updateValues = Map(":c" -> new AttributeValue().withS(newStatus))
    val request      = new UpdateItemRequest()
      .withTableName(TABLE_NAME)
      .withKey(getHashKey(channelId).asJava)
      .withUpdateExpression("set connectionStatus = :c")
      .withExpressionAttributeValues(updateValues.asJava)
    try {
      val response = dynamoDb.updateItem(request)
      logger.info(
          "Updating status for {} responded with {}",
          channelId,
          response.getSdkHttpMetadata.getHttpStatusCode
      )
    } catch {
      case e: Exception =>
        logger.warn(
            "Error occurred when trying to update status for {}",
            channelId,
            e
        )
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
  ): Unit = {
    val connection   = getConnectionDetailsById(channelId)
    connection.updateTwitchAccessToken(accessToken)
    val updateValues = Map(":s" -> new AttributeValue().withS(connection.sess))
    val request      = new UpdateItemRequest()
      .withTableName(TABLE_NAME)
      .withKey(getHashKey(channelId).asJava)
      .withUpdateExpression("set sess = :s")
      .withExpressionAttributeValues(updateValues.asJava)
    try {
      val response = dynamoDb.updateItem(request)
      logger.info(
          "Updating session object for {} responded with: {}",
          channelId,
          response.getSdkHttpMetadata.getHttpStatusCode
      )
    } catch {
      case e: Exception =>
        logger.warn(
            "Error occurred when trying to update Twitch token for {}",
            channelId,
            e
        )
    }
  }

  /**
   * Checks if a record with the given ID exists.
   * @param channelId the Twitch channel ID
   * @return true if the channel ID exists, false otherwise
   */
  override def hasConnectionDetails(channelId: String): Boolean = {
    val request = new GetItemRequest()
      .withTableName(TABLE_NAME)
      .withConsistentRead(true)
      .withKey(getHashKey(channelId).asJava)
      .withAttributesToGet("channelId") // only get one attribute for efficiency
    dynamoDb.getItem(request).getItem != null
  }

  override def stop(): Unit =
    dynamoDb.shutdown()

  private def init(): Unit = {
    logger.info("Initializing connection data store")
    val tables =
      try dynamoDb
        .listTables()
        .getTableNames
        .asScala
        .filter(_.contains(TABLE_NAME))
      catch {
        case e: Exception =>
          logger.warn("Error occurred when trying to list tables", e)
          throw e
      }
    if (tables.isEmpty) {
      // need to create the table since it doesn't exist yet.
      val request  = new CreateTableRequest()
        .withTableName(TABLE_NAME)
        .withAttributeDefinitions(
            new AttributeDefinition("channelId", ScalarAttributeType.S)
        )
        .withKeySchema(new KeySchemaElement("channelId", KeyType.HASH))
        .withBillingMode(BillingMode.PAY_PER_REQUEST)
      val response = dynamoDb.createTable(request)
      logger.info(
          "Created new DynamoDB table {} responded with status {}",
          TABLE_NAME,
          response.getSdkHttpMetadata.getHttpStatusCode
      )
    } else {
      logger.info("DynamoDB table for connections already exists")
    }
  }

  private def getHashKey(channelId: String): Map[String, AttributeValue] = Map(
      "channelId" -> new AttributeValue().withS(channelId)
  )
}

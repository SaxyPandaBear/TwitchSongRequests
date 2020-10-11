package com.github.saxypandabear.songrequests.ddb.model

import com.amazonaws.services.dynamodbv2.document.Item
import com.fasterxml.jackson.annotation.{JsonCreator, JsonIgnore, JsonProperty}
import com.fasterxml.jackson.databind.annotation.JsonDeserialize
import com.fasterxml.jackson.databind.node.ObjectNode
import com.github.saxypandabear.songrequests.json.ConnectionDeserializer
import com.github.saxypandabear.songrequests.util.JsonUtil.objectMapper

/**
 * Should look something like this:
 * {
 * "connectionStatus": {
 * "S": "active"
 * },
 * "expires": {
 * "N": "1600617110"
 * },
 * "type": {
 * "S": "connect-session"
 * },
 * "channelId": {
 * "S": "577228983"
 * },
 * "sess": {
 * "S": "{\"cookie\":{
 * \"originalMaxAge\":null,
 * \"expires\":null,
 * \"httpOnly\":true,
 * \"path\":\"/\"
 * },
 * \"accessKeys\":{
 * \"twitchToken\":{
 * \"access_token\":\"abcdefghijklmnop\",
 * \"refresh_token\":\"abcdefghijklmnop\",
 * \"token_type\":\"bearer\"
 * },
 * \"spotifyToken\":{
 * \"access_token\":\"abcdefghijklmnop\",
 * \"refresh_token\":\"abcdefghijklmnop\",
 * \"token_type\":\"Bearer\",
 * \"scope\":\"user-modify-playback-state user-read-playback-state\"
 * }
 * }
 * }"
 * }
 * }
 *
 * Note: we only deal with Twitch connections in this module, so we can cater methods to specifically get the Twitch
 * credentials, and ignore the Spotify credentials.
 *
 * This is a POJO that represents the data that is persisted to DynamoDB. This does not deal
 * with DynamoDB itself.
 */
@JsonDeserialize(using = classOf[ConnectionDeserializer])
case class Connection(
    @JsonProperty("channelId") channelId: String,
    @JsonProperty("connectionStatus") connectionStatus: String,
    @JsonProperty("expires") expires: Long,
    @JsonProperty("type") `type`: String,
    @JsonProperty("sess") var sess: String
) {
  // the refresh token doesn't change. when we parse the session object the
  // first time,
  // we can store it here so we can reference it instead of parsing the object
  // again
  private var refreshToken: String = _

  @JsonCreator
  def this() {
    this("", "", 0L, "", "")
  }

  /**
   * Since the session can change state at any point, we have to parse it every time.
   * Parse the session JSON string to get the Twitch access token.
   * @return the Twitch access token associated with this channel ID
   */
  @JsonIgnore
  def retrieveAccessToken(): String =
    extractAccessToken(extractTwitchFromSession())

  private def extractAccessToken(twitchToken: ObjectNode): String =
    twitchToken.get("access_token").asText()

  /**
   * The refresh token value doesn't change. We can/should cache this value so we
   * don't have to parse the session object every time we need to refresh - just
   * the first time.
   * @return the Twitch refresh token associated with this channel ID
   */
  @JsonIgnore
  def retrieveRefreshToken(): String = {
    if (refreshToken == null) {
      refreshToken = extractRefreshToken(extractTwitchFromSession())
    }
    refreshToken
  }

  /**
   * Parse the JSON object that represents the session state, and return a Jackson
   * ObjectNode. We cast it to an ObjectNode because a JsonNode does not provide a
   * way to update the value in the tree (see `setAccessToken`).
   * @return the parsed JSON object
   */
  private def extractTwitchFromSession(): ObjectNode =
    objectMapper
      .readTree(sess)
      .get("accessKeys")
      .get("twitchToken")
      .asInstanceOf[ObjectNode]

  private def extractRefreshToken(twitchToken: ObjectNode): String =
    twitchToken.get("refresh_token").asText()

  /**
   * Assuming that we retrieve a new access token from the authentication server,
   * we need to update our session object that we have in memory.
   * This should not update DynamoDB. We should let the ConnectionDataStore deal with
   * that.
   *
   * Note that we don't need to update any internal variable other than the session
   * JSON string, because the accessToken is parsed from the session object every time,
   * and is not cached.
   * @param token new access token that is retrieved from the server
   */
  @JsonIgnore
  def updateAccessToken(token: String): Unit = {
    val sessionObject = objectMapper.readTree(sess).asInstanceOf[ObjectNode]

    // shouldn't use the `extractTwitchFromSession` method because
    // it returns an object that is nested deeper than the root session object.
    sessionObject
      .get("accessKeys")
      .get("twitchToken")
      .asInstanceOf[ObjectNode]
      .put("access_token", token)

    // now we need to write this back to the session variable
    sess = objectMapper.writeValueAsString(sessionObject)
  }

  /**
   * Convert this Connection object into a DynamoDB interface so that we can persist it
   * to DynamoDB
   * @return
   */
  @JsonIgnore
  def toItem: Item =
    new Item()
      .withPrimaryKey("channelId", channelId)
      .withString("connectionStatus", connectionStatus)
      .withNumber("expires", expires)
      .withString("type", `type`)
      .withString("sess", sess)
}

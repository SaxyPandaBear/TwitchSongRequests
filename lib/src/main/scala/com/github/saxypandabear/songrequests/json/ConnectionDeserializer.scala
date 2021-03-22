package com.github.saxypandabear.songrequests.json

import com.fasterxml.jackson.core.JsonParser
import com.fasterxml.jackson.databind.deser.std.StdDeserializer
import com.fasterxml.jackson.databind.node.ObjectNode
import com.fasterxml.jackson.databind.{DeserializationContext, JsonNode}
import com.github.saxypandabear.songrequests.ddb.model.Connection

/**
 * Custom Jackson deserializer to deal with the Connection object.
 * TODO: It's possible that this is only needed for internal testing where we create
 *       the Connection object from test JSON files.
 */
class ConnectionDeserializer
    extends StdDeserializer[Connection](classOf[Connection]) {
  override def deserialize(
      parser: JsonParser,
      context: DeserializationContext
  ): Connection = {
    val parsed           =
      parser.getCodec.readTree[JsonNode](parser).asInstanceOf[ObjectNode]
    // build a connection object by getting all of the relevant information from
    // the JSON object
    val channelId        = parsed.get("channelId").get("S").asText()
    val connectionStatus = parsed.get("connectionStatus").get("S").asText()
    val expires          = parsed.get("expires").get("N").asLong()
    val connectionType   = parsed.get("type").get("S").asText()
    val session          = parsed.get("sess").get("S").asText()

    Connection(channelId, connectionStatus, expires, connectionType, session)
  }
}

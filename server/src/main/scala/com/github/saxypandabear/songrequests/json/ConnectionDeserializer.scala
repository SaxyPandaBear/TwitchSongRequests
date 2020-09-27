package com.github.saxypandabear.songrequests.json

import com.fasterxml.jackson.core.JsonParser
import com.fasterxml.jackson.databind.{DeserializationContext, JsonNode}
import com.fasterxml.jackson.databind.deser.std.StdDeserializer
import com.github.saxypandabear.songrequests.ddb.model.Connection

/**
 * Custom Jackson deserializer to deal with the Connection object.
 * TODO: It's possible that this is only needed for internal testing where we create
 *       the Connection object from test JSON files.
 */
class ConnectionDeserializer extends StdDeserializer[Connection](classOf[Connection]){
    override def deserialize(parser: JsonParser, context: DeserializationContext): Connection = {
        val parsed = parser.getCodec.readTree[JsonNode](parser)
        // build a connection object
    }
}

package com.github.saxypandabear.songrequests.oauth

case class OauthRequest(clientId: String, clientSecret: String, refreshToken: String) {
    def toJsonString: String =
        f"""
           |{
           |    "grant_type": "refresh_token",
           |    "refresh_token": "$refreshToken",
           |    "client_id": "$clientId",
           |    "client_secret": "$clientSecret"
           |}
           |""".stripMargin
}

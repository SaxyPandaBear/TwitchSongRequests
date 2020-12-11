package com.github.saxypandabear.songrequests.util

import org.eclipse.jetty.client.HttpClient

object HttpUtil {
  def withAutoClosingHttpClient[T](fn: HttpClient => T): T = {
    val httpClient = new HttpClient()
    httpClient.start()
    try fn(httpClient)
    finally httpClient.stop()
  }
}

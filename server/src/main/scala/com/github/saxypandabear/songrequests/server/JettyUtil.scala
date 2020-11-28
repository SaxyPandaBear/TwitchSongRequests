package com.github.saxypandabear.songrequests.server

import org.eclipse.jetty.server.Server
import org.eclipse.jetty.servlet.{ServletContextHandler, ServletHolder}
import org.glassfish.jersey.internal.inject.AbstractBinder
import org.glassfish.jersey.server.ResourceConfig
import org.glassfish.jersey.servlet.ServletContainer

object JettyUtil {

  def build(port: Int, binder: AbstractBinder = null): Server = {
    val server = new Server(port)
    server.setStopAtShutdown(true)
    server.setStopTimeout(0)

    val ctx = new ServletContextHandler(ServletContextHandler.NO_SESSIONS)
    ctx.setContextPath("/")

    val resourceConfig = new ResourceConfig()
    if (binder != null) {
      resourceConfig.register(binder, 1)
    }

    resourceConfig.packages("com.github.saxypandabear.songrequests.server")
    val servletContainer = new ServletContainer(resourceConfig)
    val servletHolder    = new ServletHolder(servletContainer)
    ctx.addServlet(servletHolder, "/*")

    server.setHandler(ctx)
    server
  }
}

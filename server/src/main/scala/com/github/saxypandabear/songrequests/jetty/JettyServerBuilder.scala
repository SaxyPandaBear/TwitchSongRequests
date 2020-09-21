package com.github.saxypandabear.songrequests.jetty

import org.eclipse.jetty.server.Server
import org.eclipse.jetty.servlet.{ServletContextHandler, ServletHolder}
import org.glassfish.jersey.server.ResourceConfig
import org.glassfish.jersey.servlet.ServletContainer

object JettyServerBuilder {

    def build(port: Int): Server = {
        val server = new Server(port)
        server.setStopAtShutdown(true)
        server.setStopTimeout(0)

        val ctx = new ServletContextHandler(ServletContextHandler.NO_SESSIONS)
        ctx.setContextPath("/")

        val resourceConfig = new ResourceConfig()
        resourceConfig.packages("com.github.saxypandabear.songrequests.jetty")
        val servletContainer = new ServletContainer(resourceConfig)
        val servletHolder = new ServletHolder(servletContainer)
        ctx.addServlet(servletHolder, "/*")

        server.setHandler(ctx)
        server
    }
}

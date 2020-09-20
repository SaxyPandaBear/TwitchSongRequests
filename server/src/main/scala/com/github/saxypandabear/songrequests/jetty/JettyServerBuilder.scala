package com.github.saxypandabear.songrequests.jetty

import org.eclipse.jetty.server.Server
import org.eclipse.jetty.servlet.ServletContextHandler

object JettyServerBuilder {

    def build(port: Int): Server = {
        val server = new Server(port)

        val ctx = new ServletContextHandler(ServletContextHandler.NO_SESSIONS)
        ctx.setContextPath("/")

        server.setHandler(ctx)

        server
    }
}

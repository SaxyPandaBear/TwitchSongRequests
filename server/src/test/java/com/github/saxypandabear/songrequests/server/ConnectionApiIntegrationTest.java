package com.github.saxypandabear.songrequests.server;

import cloud.localstack.LocalstackTestRunner;
import cloud.localstack.docker.annotation.LocalstackDockerProperties;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.github.saxypandabear.songrequests.server.model.Channel;
import com.github.saxypandabear.songrequests.util.JsonUtil;
import com.github.saxypandabear.songrequests.util.ProjectProperties;
import com.github.saxypandabear.songrequests.websocket.lib.WebSocketTestingUtil;
import io.restassured.RestAssured;
import io.restassured.http.ContentType;
import org.eclipse.jetty.server.Server;
import org.junit.After;
import org.junit.Before;
import org.junit.BeforeClass;
import org.junit.Test;
import org.junit.runner.RunWith;

import java.nio.file.Files;
import java.nio.file.Path;
import java.util.Random;

import static org.junit.Assert.*;

@RunWith(LocalstackTestRunner.class)
@LocalstackDockerProperties(services = {"cloudwatch"})
public class ConnectionApiIntegrationTest {

    private final Random random = new Random(System.currentTimeMillis());

    private Path tempPropertiesFilePath;
    private int apiPort;
    private int socketPort;
    private Server socketServer;

    @BeforeClass
    public static void beforeAll() {
        RestAssured.enableLoggingOfRequestAndResponseIfValidationFails();
        RestAssured.useRelaxedHTTPSValidation();
    }

    @Before
    public void setUp() {
        apiPort = randomPort(5000);
        socketPort = randomPort(8000);

        socketServer = WebSocketTestingUtil.build(socketPort);

        ProjectProperties properties = new ProjectProperties();
        properties.setValue("port", Integer.toString(apiPort));
        properties.setValue("cloudwatch.url", "http://localhost:4566");
        properties.setValue("twitch.port", Integer.toString(socketPort));
        tempPropertiesFilePath = properties.toTemporaryFile("integration-test");
        Main.main(new String[]{tempPropertiesFilePath.toString()});
    }

    @After
    public void cleanUp() throws Exception {
        Main.stop();
        socketServer.stop();
        Files.deleteIfExists(tempPropertiesFilePath);
    }

    @Test
    public void pingShouldRespondWithPong() {
        String responseBody = RestAssured
                .get(String.format("http://localhost:%d/api/ping", apiPort))
                .then()
                .extract()
                .body()
                .asString();
        assertEquals("pong", responseBody);
    }

    @Test
    public void connectResponse() throws JsonProcessingException {
        String id = "12345";
        Channel channel = Channel.apply(id);

        String response = RestAssured
                .given()
                .contentType(ContentType.JSON)
                .accept(ContentType.TEXT)
                .body(JsonUtil.objectMapper().writeValueAsString(channel))
                .post(String.format("http://localhost:%d/api/connect", apiPort))
                .then()
                .assertThat()
                .statusCode(201)
                .and()
                .extract()
                .body()
                .asString();
        assertEquals(String.format("Initiated connection to channel %s", id), response);

        // peek into the internal state of the orchestrator to verify that it connected.
        assertTrue(
                "ID should be visible in the orchestrator",
                Main.orchestrator().connectionsToClients().exists(tuple -> tuple._2.contains(id))
        );
    }

    @Test
    public void disconnectResponse() throws JsonProcessingException {
        String id = "12345";
        Channel channel = Channel.apply(id);

        // first, need to connect to the server
        String response = RestAssured
                .given()
                .contentType(ContentType.JSON)
                .accept(ContentType.TEXT)
                .body(JsonUtil.objectMapper().writeValueAsString(channel))
                .post(String.format("http://localhost:%d/api/connect", apiPort))
                .then()
                .assertThat()
                .statusCode(201)
                .and()
                .extract()
                .body()
                .asString();
        assertEquals(String.format("Initiated connection to channel %s", id), response);

        // peek into the internal state of the orchestrator to verify that it connected.
        assertTrue(
                "ID should be visible in the orchestrator",
                Main.orchestrator().connectionsToClients().exists(tuple -> tuple._2.contains(id))
        );

        RestAssured
                .given()
                .contentType(ContentType.JSON)
                .accept(ContentType.ANY)
                .pathParam("channel", id)
                .put(String.format("http://localhost:%d/api/disconnect/{channel}", apiPort))
                .then()
                .assertThat()
                .statusCode(204);

        // peek into the internal state to make sure that the disconnected
        // channel doesn't show up.
        assertFalse(
                "ID should be fully disconnected",
                Main.orchestrator().connectionsToClients().exists(tuple -> tuple._2.contains(id))
        );
    }

    private int randomPort(int base) {
        return random.nextInt(1000) + base;
    }
}

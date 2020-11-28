package com.github.saxypandabear.songrequests.server;

import cloud.localstack.LocalstackTestRunner;
import cloud.localstack.docker.annotation.LocalstackDockerProperties;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.github.saxypandabear.songrequests.server.model.Channel;
import com.github.saxypandabear.songrequests.util.JsonUtil;
import com.github.saxypandabear.songrequests.util.ProjectProperties;
import io.restassured.RestAssured;
import io.restassured.http.ContentType;
import org.junit.After;
import org.junit.Before;
import org.junit.BeforeClass;
import org.junit.Test;
import org.junit.runner.RunWith;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.Random;

import static org.junit.Assert.assertEquals;

@RunWith(LocalstackTestRunner.class)
@LocalstackDockerProperties(services = {"cloudwatch"})
public class ConnectionApiIntegrationTest {

    private final Random random = new Random(System.currentTimeMillis());

    private Path tempPropertiesFilePath;
    private int port;

    @BeforeClass
    public static void beforeAll() {
        RestAssured.enableLoggingOfRequestAndResponseIfValidationFails();
        RestAssured.useRelaxedHTTPSValidation();
    }

    @Before
    public void setUp() {
        port = randomPort();

        ProjectProperties properties = new ProjectProperties();
        properties.setValue("port", Integer.toString(port));
        tempPropertiesFilePath = properties.toTemporaryFile("integration-test");
        Main.main(new String[]{tempPropertiesFilePath.toString()});
    }

    @After
    public void cleanUp() throws IOException {
        Main.stop();
        Files.deleteIfExists(tempPropertiesFilePath);
    }

    @Test
    public void pingShouldRespondWithPong() {
        String responseBody = RestAssured
                .get(String.format("http://localhost:%d/api/ping", port))
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
                .post(String.format("http://localhost:%d/api/connect", port))
                .then()
                .assertThat()
                .statusCode(201)
                .and()
                .extract()
                .body()
                .asString();
        assertEquals(String.format("Initiated connection to channel %s", id), response);
    }

    @Test
    public void disconnectResponse() {
        String id = "12345";
        RestAssured
                .given()
                .contentType(ContentType.JSON)
                .accept(ContentType.ANY)
                .pathParam("channel", id)
                .put(String.format("http://localhost:%d/api/disconnect/{channel}", port))
                .then()
                .assertThat()
                .statusCode(204);
    }

    private int randomPort() {
        return random.nextInt(1000) + 5000;
    }
}

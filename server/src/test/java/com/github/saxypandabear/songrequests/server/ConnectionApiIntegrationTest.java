package com.github.saxypandabear.songrequests.server;

import cloud.localstack.Localstack;
import cloud.localstack.LocalstackTestRunner;
import cloud.localstack.docker.annotation.LocalstackDockerProperties;
import com.amazonaws.client.builder.AwsClientBuilder;
import com.amazonaws.services.dynamodbv2.AmazonDynamoDB;
import com.amazonaws.services.dynamodbv2.AmazonDynamoDBClientBuilder;
import com.amazonaws.services.dynamodbv2.model.PutItemResult;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.github.saxypandabear.songrequests.ddb.model.Connection;
import com.github.saxypandabear.songrequests.server.model.Channel;
import com.github.saxypandabear.songrequests.util.JsonUtil;
import com.github.saxypandabear.songrequests.util.ProjectProperties;
import com.github.saxypandabear.songrequests.websocket.lib.WebSocketTestingUtil$;
import io.restassured.RestAssured;
import io.restassured.http.ContentType;
import org.eclipse.jetty.server.Server;
import org.junit.After;
import org.junit.Before;
import org.junit.BeforeClass;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.Objects;
import java.util.Random;
import java.util.UUID;

import static org.junit.Assert.*;

@RunWith(LocalstackTestRunner.class)
@LocalstackDockerProperties(services = {"cloudwatch", "sqs", "dynamodb"})
public class ConnectionApiIntegrationTest {

    private static final Logger logger = LoggerFactory.getLogger(ConnectionApiIntegrationTest.class);
    private static final String connectionTemplateJsonPath = "test-json/connection-active.json";
    private static final Connection template = readTemplateConnectionObject();

    private final Random random = new Random(System.currentTimeMillis());

    private static AmazonDynamoDB ddb;

    private Path tempPropertiesFilePath;
    private int apiPort;
    private int socketPort;
    private Server socketServer;

    @BeforeClass
    public static void beforeAll() {
        RestAssured.enableLoggingOfRequestAndResponseIfValidationFails();
        RestAssured.useRelaxedHTTPSValidation();

        ddb = AmazonDynamoDBClientBuilder
                .standard()
                .withEndpointConfiguration(
                        new AwsClientBuilder.EndpointConfiguration(
                                Localstack.INSTANCE.getEndpointSQS(),
                                Localstack.getDefaultRegion()
                        ))
                .build();
    }

    @Before
    public void setUp() throws Exception {
        apiPort = randomPort(5000);
        socketPort = randomPort(8000);

        // Java doesn't play nicely with Scala object singletons
        socketServer = WebSocketTestingUtil$.MODULE$.build(socketPort);
        socketServer.start();

        ProjectProperties properties = new ProjectProperties();
        properties.setValue("env", "integration_test");
        properties.setValue("port", Integer.toString(apiPort));
        properties.setValue("client.id", "foo");
        properties.setValue("client.secret", "bar");
        properties.setValue("twitch.refresh.uri", String.format("http://localhost:%d", socketPort));
        properties.setValue("cloudwatch.url", "http://localhost:4566");
        properties.setValue("sqs.url", "http://localhost:4566");
        properties.setValue("dynamodb.url", "http://localhost:4566");
        properties.setValue("twitch.port", Integer.toString(socketPort));
        tempPropertiesFilePath = properties.toTemporaryFile("integration-test");
        Main.main(new String[]{tempPropertiesFilePath.toString()});
    }

    @After
    public void tearDown() throws Exception {
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
        String id = putNewConnection();
        successfullyConnect(id);
    }

    @Test
    public void disconnectResponse() throws JsonProcessingException {
        String id = putNewConnection();

        // first, need to connect to the server
        successfullyConnect(id);

        // after confirming that we are connected, disconnect from the server.
        successfullyDisconnect(id);
    }

    private void successfullyConnect(String channelId) throws JsonProcessingException {
        Channel channel = Channel.apply(channelId);
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
        assertEquals(String.format("Initiated connection to channel %s", channelId), response);

        // peek into the internal state of the orchestrator to verify that it connected.
        assertTrue(
                "ID should be visible in the orchestrator",
                Main.orchestrator().connectionsToClients().exists(tuple -> tuple._2.contains(channelId))
        );
    }

    private void successfullyDisconnect(String channelId) {
        RestAssured
                .given()
                .contentType(ContentType.JSON)
                .accept(ContentType.ANY)
                .pathParam("channel", channelId)
                .put(String.format("http://localhost:%d/api/disconnect/{channel}", apiPort))
                .then()
                .assertThat()
                .statusCode(204);

        // peek into the internal state to make sure that the disconnected
        // channel doesn't show up.
        assertFalse(
                "ID should be fully disconnected",
                Main.orchestrator().connectionsToClients().exists(tuple -> tuple._2.contains(channelId))
        );
    }

    private static Connection readTemplateConnectionObject() {
        try {
            return JsonUtil
                    .objectMapper()
                    .readValue(Objects
                                    .requireNonNull(ConnectionApiIntegrationTest.class
                                            .getClassLoader()
                                            .getResourceAsStream(connectionTemplateJsonPath)),
                            Connection.class
                    );
        } catch (IOException e) {
            e.printStackTrace();
            throw new RuntimeException("Failed to parse template JSON", e);
        }
    }

    /**
     * Same implementation as the RotatingPort test trait, but Java doesn't
     * play nicely with the Scala trait.
     *
     * @param base base port
     * @return base + random([0, 1000))
     */
    private int randomPort(int base) {
        return random.nextInt(1000) + base;
    }

    /**
     * Need to set up an existing connection. this component is not
     * intended to insert records in to the DynamoDB table. It should
     * just be reading and updating records. As such, we need to simulate
     * this properly by initializing the tests with a Connection.
     *
     * @return the channel ID for the connection that is written to DynamoDB
     */
    private String putNewConnection() {
        String channelId = UUID.randomUUID().toString();
        Connection connection = template.copy(channelId, template.connectionStatus(), template.expires(), template.type(), template.sess());
        PutItemResult result = ddb.putItem("connections", connection.toJavaValueMap());
        logger.info("Created new connection record with ID {}, responded with status code: {}", channelId, result.getSdkHttpMetadata().getHttpStatusCode());
        return channelId;
    }
}

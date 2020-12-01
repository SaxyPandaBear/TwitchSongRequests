package com.github.saxypandabear.songrequests.ddb;

import cloud.localstack.Localstack;
import cloud.localstack.LocalstackTestRunner;
import cloud.localstack.docker.annotation.LocalstackDockerProperties;
import com.amazonaws.client.builder.AwsClientBuilder;
import com.amazonaws.services.dynamodbv2.AmazonDynamoDB;
import com.amazonaws.services.dynamodbv2.AmazonDynamoDBClientBuilder;
import com.amazonaws.services.dynamodbv2.model.PutItemRequest;
import com.amazonaws.services.dynamodbv2.model.ResourceNotFoundException;
import com.github.saxypandabear.songrequests.ddb.model.Connection;
import com.github.saxypandabear.songrequests.util.JsonUtil;
import org.junit.After;
import org.junit.Before;
import org.junit.Rule;
import org.junit.Test;
import org.junit.rules.Timeout;
import org.junit.runner.RunWith;

import java.io.IOException;
import java.util.Objects;
import java.util.concurrent.TimeUnit;

import static org.junit.Assert.*;

/**
 * Localstack integration tests for DynamoDB implementation of data store.
 * Written in Java because Scalatest and Gradle do not play nicely with
 * incorporating a separate runner for Localstack.
 */
@RunWith(LocalstackTestRunner.class)
@LocalstackDockerProperties(services = {"dynamodb"})
public class DynamoDbConnectionDataStoreIntegrationTest {

    private static final String connectionTemplateJsonPath = "test-json/connection-active.json";

    @Rule
    public Timeout globalTimeout = new Timeout(5, TimeUnit.SECONDS);

    private DynamoDbConnectionDataStore dataStore;
    private AmazonDynamoDB ddb;

    @Before
    public void setUp() {
        ddb = AmazonDynamoDBClientBuilder
                .standard()
                .withEndpointConfiguration(
                        new AwsClientBuilder.EndpointConfiguration(
                                Localstack.INSTANCE.getEndpointSQS(),
                                Localstack.getDefaultRegion()
                        ))
                .build();
        dataStore = new DynamoDbConnectionDataStore(ddb);
    }

    // not dropping the table, so all of the items created in the tests will
    // exist until the test completes, and the container is torn down.
    @After
    public void tearDown() {
        dataStore.stop();
    }

    @Test
    public void getConnectionThatExists() {
        String channelId = createNewDbEntry();

        Connection connection = dataStore.getConnectionDetailsById(channelId);
        assertEquals(channelId, connection.channelId());
        assertEquals(9876543210L, connection.expires());
        assertEquals("active", connection.connectionStatus());
    }

    @Test(expected = ResourceNotFoundException.class)
    public void getConnectionThatDoesNotExist() {
        String channelId = Long.toString(System.currentTimeMillis());
        dataStore.getConnectionDetailsById(channelId);
    }

    @Test
    public void tableHasConnection() {
        String channelId = createNewDbEntry();
        assertTrue(dataStore.hasConnectionDetails(channelId));
    }

    @Test
    public void tableDoesNotHaveConnection() {
        String channelId = Long.toString(System.currentTimeMillis());
        assertFalse(dataStore.hasConnectionDetails(channelId));
    }

    @Test
    public void updateTwitchOAuthTokenInRecord() {
        String channelId = createNewDbEntry();
        String accessToken = "some-new-token";

        // make sure that this isn't a false positive test by checking
        // the original record
        Connection connection = dataStore.getConnectionDetailsById(channelId);
        assertNotEquals(accessToken, connection.twitchAccessToken());

        dataStore.updateTwitchOAuthToken(channelId, accessToken);

        connection = dataStore.getConnectionDetailsById(channelId);
        assertEquals(accessToken, connection.twitchAccessToken());
    }

    @Test(expected = ResourceNotFoundException.class)
    public void updateTwitchOAuthTokenInRecordDoesNotExist() {
        String channelId = Long.toString(System.currentTimeMillis());
        String accessToken = "some-new-token";
        dataStore.updateTwitchOAuthToken(channelId, accessToken);
    }

    @Test
    public void updateConnectionStatusInRecord() {
        String channelId = createNewDbEntry();
        String updatedStatus = "inactive";

        // make sure that this isn't a false positive test by checking
        // the original record
        Connection connection = dataStore.getConnectionDetailsById(channelId);
        assertNotEquals(updatedStatus, connection.connectionStatus());

        dataStore.updateConnectionStatus(channelId, updatedStatus);

        connection = dataStore.getConnectionDetailsById(channelId);
        assertEquals(updatedStatus, connection.connectionStatus());
    }

    /**
     * Read in the template JSON for a connection object, replace the ID with
     * something unique, write the record to DynamoDB, and return the primary
     * key.
     *
     * @return channel ID of the object created
     */
    private String createNewDbEntry() {
        // by using the current timestamp as a string, we ensure uniqueness
        // across tests.
        String channelId = Long.toString(System.currentTimeMillis());
        Connection connection;
        try {
            connection = JsonUtil
                    .objectMapper()
                    .readValue(Objects
                                    .requireNonNull(getClass()
                                            .getClassLoader()
                                            .getResourceAsStream(connectionTemplateJsonPath)),
                            Connection.class
                    );
        } catch (IOException e) {
            e.printStackTrace();
            throw new RuntimeException("Failed to parse template JSON", e);
        }
        // no setter, so just get a new Connection object with the updated channelId
        connection = new Connection(
                channelId,
                connection.connectionStatus(),
                connection.expires(),
                connection.type(),
                connection.sess()
        );

        PutItemRequest request = new PutItemRequest()
                .withTableName(dataStore.TABLE_NAME())
                .withItem(connection.toJavaValueMap());
        ddb.putItem(request);

        return channelId;
    }
}

package com.github.saxypandabear.songrequests.queue;

import cloud.localstack.Localstack;
import cloud.localstack.LocalstackTestRunner;
import cloud.localstack.docker.annotation.LocalstackDockerProperties;
import com.amazonaws.client.builder.AwsClientBuilder;
import com.amazonaws.services.cloudwatch.AmazonCloudWatch;
import com.amazonaws.services.cloudwatch.AmazonCloudWatchClientBuilder;
import com.amazonaws.services.sqs.AmazonSQS;
import com.amazonaws.services.sqs.AmazonSQSClientBuilder;
import com.amazonaws.services.sqs.model.Message;
import com.amazonaws.services.sqs.model.PurgeQueueRequest;
import com.amazonaws.services.sqs.model.ReceiveMessageResult;
import com.github.saxypandabear.songrequests.metric.CloudWatchMetricCollector;
import org.junit.After;
import org.junit.Before;
import org.junit.Rule;
import org.junit.Test;
import org.junit.rules.Timeout;
import org.junit.runner.RunWith;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.TimeUnit;

import static org.junit.Assert.*;

/**
 * Localstack integration tests for SQS implementation of song queue class.
 * Written in Java because Scalatest and Gradle do not play nicely with
 * incorporating a separate runner for Localstack.
 */
@RunWith(LocalstackTestRunner.class)
@LocalstackDockerProperties(services = {"sqs", "cloudwatch"})
public class SQSSongQueueIntegrationTest {

    // note: can't make the global timeout too small because the first test
    //       has to create the SQS queue, which adds a considerable amount
    //       of overhead to the test that goes first (roughly 2 seconds
    //       observed).
    @Rule
    public Timeout globalTimeout = new Timeout(5, TimeUnit.SECONDS);

    private static final String CHANNEL_ID = "abc123";

    private SQSSongQueue songQueue;
    private AmazonCloudWatch cloudWatch;
    private AmazonSQS sqs;
    private ExecutorService threadPool;

    @Before
    public void setUp() {
        cloudWatch = AmazonCloudWatchClientBuilder
                .standard()
                .withEndpointConfiguration(
                        new AwsClientBuilder.EndpointConfiguration(
                                Localstack.INSTANCE.getEndpointCloudWatch(),
                                Localstack.getDefaultRegion()
                        ))
                .build();

        sqs = AmazonSQSClientBuilder
                .standard()
                .withEndpointConfiguration(
                        new AwsClientBuilder.EndpointConfiguration(
                                Localstack.INSTANCE.getEndpointSQS(),
                                Localstack.getDefaultRegion()
                        ))
                .build();
        threadPool = Executors.newFixedThreadPool(5);
        songQueue = new SQSSongQueue(sqs, new CloudWatchMetricCollector(cloudWatch, threadPool));
    }

    @After
    public void cleanUp() {
        // drop all of the existing messages from testing to prep for the next test.
        sqs.purgeQueue(new PurgeQueueRequest().withQueueUrl(songQueue.getQueueUrl()));
        songQueue.stop();
        cloudWatch.shutdown();
        threadPool.shutdown();
    }

    @Test
    public void queueSong() {

        String song = "some-song";
        songQueue.queue(CHANNEL_ID, song);

        // after the song is queued, we should be able to read the message from
        // the SQS queue
        Message foundMessage = null;
        while (foundMessage == null) {
            ReceiveMessageResult response = sqs.receiveMessage(songQueue.getQueueUrl());
            if (!response.getMessages().isEmpty()) {
                // because the default configuration is to only fetch one
                // message at a time, we are okay just grabbing the first
                // message.
                foundMessage = response.getMessages().get(0);
            }
        }

        assertNotNull("A message should have been read from the queue before timing out", foundMessage);
        assertEquals(song, foundMessage.getBody());
        assertTrue("Message attributes should have the 'channelId' attribute", foundMessage.getMessageAttributes().containsKey("channelId"));
        assertEquals(CHANNEL_ID, foundMessage.getMessageAttributes().get("channelId").getStringValue());
    }

    @Test
    public void queueMultipleSongs() {
        // defined as an ArrayList with an Arrays.asList() so that we can mutate
        List<String> songs = new ArrayList<>(Arrays.asList("song1", "song2", "song3", "song4", "song5"));

        for (String song : songs) {
            songQueue.queue(CHANNEL_ID, song);
        }

        // exhaust all of the songs that were queued up. if the test times out
        // before this is done checking all the messages, either SQS is slow,
        // or we didn't receive all the messages (assume the latter).
        while (!songs.isEmpty()) {
            ReceiveMessageResult response = sqs.receiveMessage(songQueue.getQueueUrl());
            for (Message message : response.getMessages()) {
                assertTrue("Message attributes should have the 'channelId' attribute", message.getMessageAttributes().containsKey("channelId"));
                assertEquals(CHANNEL_ID, message.getMessageAttributes().get("channelId").getStringValue());

                String song = message.getBody();
                assertTrue("The song should have been something queued in this test", songs.contains(song));
                songs.remove(song);
            }
        }
    }
}

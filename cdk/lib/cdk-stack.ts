import * as cdk from '@aws-cdk/core'
import * as lambda from '@aws-cdk/aws-lambda'
import {SqsEventSource} from '@aws-cdk/aws-lambda-event-sources'
import * as sqs from '@aws-cdk/aws-sqs'

export class CdkStack extends cdk.Stack {
  constructor(scope: cdk.Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props)

    // Can add timeout, deadLetterQueue, memorySize etc. later to fine tune the created lambda
    const queueSongLambda = new lambda.Function( this, 'QueueSongLambda', {
      code: new lambda.AssetCode(`../queue-song-lambda/src`),
      handler: 'Foo.foo',
      runtime: lambda.Runtime.JAVA_8,
    })
    const songRequestsQueue = new sqs.Queue(this, 'SongRequestQueue', {
      visibilityTimeout: cdk.Duration.seconds(30), //default
      receiveMessageWaitTime: cdk.Duration.seconds(20)
    })
    const queueSongEventSource = queueSongLambda.addEventSource(new SqsEventSource(songRequestsQueue, {
      batchSize: 10, //default
    }));
  }
}

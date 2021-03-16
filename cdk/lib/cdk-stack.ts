import {Construct, Duration, Stack, StackProps} from 'monocdk'
import {AssetCode, Function, Runtime} from 'monocdk/aws-lambda'
import {SqsEventSource} from 'monocdk/aws-lambda-event-sources'
import {Queue} from 'monocdk/aws-sqs'

export class CdkStack extends Stack {
  constructor(scope: Construct, id: string, props?: StackProps) {
    super(scope, id, props)

    // Can add timeout, deadLetterQueue, memorySize etc. later to fine tune the created lambda
    const queueSongLambda = new Function( this, 'QueueSongLambda', {
      code: new AssetCode(`../queue-song-lambda/src`),
      handler: 'Foo.foo',
      runtime: Runtime.JAVA_8,
    })
    const songRequestsQueue = new Queue(this, 'SongRequestQueue', {
      visibilityTimeout: Duration.seconds(30), //default
      receiveMessageWaitTime: Duration.seconds(20)
    })
    const queueSongEventSource = queueSongLambda.addEventSource(new SqsEventSource(songRequestsQueue, {
      batchSize: 10, //default
    }));
  }
}

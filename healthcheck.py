"""
Read stdin, which is expected to be the response from the /health check from 
localstack. This should parse the JSON response, and exit successfully if, and only 
if, all of the services are "running"

Note: this expects the input to be valid. The shell script that is orchestrating the 
startup process should handle when cURL doesn't receive a response back from localstack
(such as when the service hasn't started yet).
"""
import sys
import json

response = json.load(sys.stdin)
# Healthcheck repsonse looks like this:
# response = {
#     "services": {
#         "apigateway": "starting", 
#         "cloudformation": "running", 
#         "cloudwatch": "running", 
#         "dynamodb": "running", 
#         "dynamodbstreams": "running", 
#         "ec2": "running", "es": "running", 
#         "firehose": "running", 
#         "iam": "running", 
#         "sts": "running", 
#         "kinesis": "running", 
#         "kms": "running", 
#         "lambda": "running", 
#         "logs": "running", 
#         "redshift": "running", 
#         "route53": "running", 
#         "s3": "running", 
#         "secretsmanager": "running", 
#         "ses": "running", 
#         "sns": "running", 
#         "sqs": "running", 
#         "ssm": "running", 
#         "events": "running",
#         "stepfunctions": "running", 
#         "acm": "running"
#     }
# }
services = response["services"]

for service,status in services.items():
    if status != "running":
        print(f"{service} is not running; Current status is {status}")
        exit(1)
print("All localstack services are running!")
exit(0)

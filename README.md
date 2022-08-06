# AWS Demo Project
Demo project 4 learning purposes. Use case / process: On every event `s3:ObjectCreated:*`, process the event via a SNS & SQS fan-out pattern, get queue entry processed by Lambda function (event mapping) to create a record in a DynamoDB table. Serve app via EC2 with a list of all objects and a pre-signed URL to access all files anonymously. During indexing, all table items get a expiration, so the items are listed for a limited time via EC2 app, but do not get deleted in S3.

## Requirements
- Golang 1.18
- `task` CLI tool (https://taskfile.dev)
- AWS CLI v2 installed & configured with sufficient AWS credentials

## Initial configuration
Create `.env` file:
```
KEY_PAIR=
AMI=
STACKNAME=
```
_(eu-central-1 AMZ Linux2 AMI: `ami-0c956e207f9d113d5`)_

Care: S3 bucket will be named _cloudformation-STACKNAME-bucket_. Bucket names need to be unique per AWS region, so stack names like _demo_ are very likely to fail.

`STACKNAME` can be overridden with `STACK` env variable, e.g.
```
STACK=example-stack task create
```

## Commands

### Launch
- Create stack with `task create` and wait until it's created (otherwise _cannot resolve hostname_ error)
- Deploy Go app to created instance with `task deploy`

### Update
- To update the AWS Stack, run `task update`
- To update the Go app, re-run `task deploy`

### Delete
- Empty the created S3 bucket via AWS Console
- Run `task delete`
- If S3 bucket is purged after getting `DELETE_FAILED` state because of bucket content, fire 2nd `task delete`

## Possible enhancements (additional learning ideas)
- Make use of fan-out pattern (currently unnecessary, there's only one topic subscription)
- Use SNS subscription filters, e.g. on file ending to only list safe files like images
- Make EC2 instance highly available with an ASG
- Move Go app deployment control from local to CodePipeline
- Migrate app from EC2 to ECS or EKS
- XRay
- Encryption
- Something with Lambda layers

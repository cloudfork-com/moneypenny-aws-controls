Goto the folder `lambda` to find a Makefile and resources for deployment.

#### create IAM Policy

Create a policy named `moneypenny-aws-controls-policy` using permissions as defined in `iam-policy.json`.

#### create IAM Role

Create a role named `moneypenny-aws-controls-role`:

- Trusted entity type = AWS Service
- Use case, Service = Lambda
- Policy = `moneypenny-aws-controls-policy`

#### deploy Lambda service (ARM64)

The `Makefile` has a `TIME_ZONE` environment variable you might need to change.
In the commands below, replace the ROLE arn with that of `moneypenny-aws-controls-role` you just created.

```
make compile 
make zip
ROLE=arn:aws:iam::111111111:role/moneypenny-aws-controls-role make create
```

#### add trigger for Lambda service (API Gateway)

- create a new API Gateway HTTP API

Now you have an API endpoint that you can visit to see the current Schedule.

#### define apply schedule

Using the Amazon EventBridge Scheduler you define a new schedule that targets the service.

- create a new schedule called `apply-hourly-moneypenny-aws-controls`
- set the cron expression to `5 * * * ? *` , which means run 5 minutes past every hour
- set the target to the Lambda `moneypenny-aws-controls`
- set the payload to:
```
{
    "queryStringParameters":{
        "do":"apply"
    }
}
```
- change the role name to `Amazon_EventBridge_Scheduler_LAMBDA_moneypenny_aws_controls` for better recognition when listing roles in AWS console.
## moneypenny AWS controls

Service to schedule ECS services by controlling the `desired-count` value.

![schedule](doc/schedule.png)

## cron

```
 "0 15 1-5"
 ```

 - minute, 0..59
 - hour, 0..23
 - day of week, use range "2-3" or comma separated values "1,2,3,4". Sunday=0


### AWS tag

Using a tag with key `moneypenny`, you can specify the cron expressions for both `running` and `stopped` states.

To run a service between 08:00 and 18:00 on workdays (1=Monday,5=Friday), use:
```
running=0 8 1-5. stopped=0 18 1-5.
```

To stop a service indefinitely, use:
```
stopped=0 0 0-6.
```

To run a service indefinitely either remove or empty the tag or use:
```
running=0 0 0-6.
```

To disable one or all state changes, use the comment indicator `//`:
```
// running=0 8 1-5. stopped=0 18 1-5. count=4
running=0 8 1-5. // stopped=0 18 1-5. count=4
```


### Local config

In `aws-service-plans.json` specify the services and cron expressions.
Local enabled defined plans override the onces defined in AWS.

```
[
    { 
        "service-arn": "arn:aws:ecs:eu-central-1:9111111:service/cluster/name",
        "moneypenny": "running=0 8 1-5. stopped=0 18 1-5."
    },
    { 
        "service-arn": "arn:aws:ecs:eu-central-1:9111111:service/cluster/ignore",
        "moneypenny": "running=0 8 1-5. stopped=0 18 1-5.",
        "disabled": true
    },
    { 
        "service-arn": "arn:aws:ecs:eu-central-1:9111111:service/cluster/high-load",
        "moneypenny": "running=0 7 1-5. stopped=0 22 1-5. count=2."
    }
]
```

### AWS deployment

Goto the folder `lambda` to find scripts and resources for deployment.


#### create IAM Policy

Create a policy named `moneypenny-aws-controls-policy` using permissions as defined in `iam-policy.json`.


#### create IAM Role

Create a role named `moneypenny-aws-controls-role`:

- Trusted entity type = AWS Service
- Service = Lambda
- Policy = moneypenny-aws-controls-policy


#### deploy Lambda service (ARM64)

The `Makefile` has a `TIME_ZONE` environment variable you might need to change.
In the commands below, replace the ROLE arn with that of `moneypenny-aws-controls-role`.

```
make compile 
make zip
ROLE=arn:aws:iam::111111111:role/moneypenny-aws-controls-role make create
```

#### define apply schedule

Using the Amazon EventBridge Scheduler you define a new schedule that targets the service.

- create a new schedule called `apply-hourly-moneypenny-aws-controls`
- set the cron expression to `5 * * * ? *` , which means run 5 minutes past every hour
- set the target to the Lambda `moneypenny-aws-controls`

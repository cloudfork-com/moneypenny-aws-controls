## moneypenny AWS controls

Service to schedule ECS services by controlling the `desired-count` value.

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

### Local config

In `aws-service-plans.json` specify the services and cron expressions.

```
[
    { 
        "service-arn": "arn:aws:ecs:eu-central-1:9111111:service/cluster/name",
        "moneypenny": "running=0 8 1-5. stopped=0 18 1-5."
    },
    { 
        "service-arn": "arn:aws:ecs:eu-central-1:9111111:service/cluster/high-load",
        "desired-tasks-count": 4,
        "moneypenny": "running=0 7 1-5. stopped=0 22 1-5."
    }
]
```
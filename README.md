## moneypenny AWS controls

Service to schedule ECS tasks.

- list running/stopped Fargate tasks
- set startstop event for a task

## cron

```
 "0 15 1-5"
 ```

 - minute, 0..59
 - hour, 0..23
 - day of week, use range "2-3" or comma separated values "1,2,3,4". Sunday=0


### AWS tag

Using a tag with key `moneypenny`, you can specify the cron expressions for both `running` and `stopped` states.

Example:
```
running=0 8 1-5. stopped=0 18 1-5.
```
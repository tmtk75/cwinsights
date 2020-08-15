# README
`cwinsights` is a utility command for CloudWatch Logs Insights for the purpose of bulky query.


## Motivation
AWS Step Functions consists of several Lambda, ECS Task, etc.
Their logs are basically stored into CloudWatch Logs and some log groups are created.
Sometimes I want to query a pattern in all such log groups in one-shot.
This command supports it.


## Getting Started
`bulk` is to use the purpose. It ends synchronously.
```
$ cat > groups
/aws/lambda/foo1
/aws/lambda/foo2
/aws/lambda/foo3

$ cwinsights bulk groups \
    --since 1h \
    --query-string 'fields @message | filter @message =~ /INFO/'
...
```

Simply `query` is supported. `--group-name` gives log group name.
It's also convenient synchronously to print results.
```
$ cwinsights query \
    --since 1h \
    --query-string "INFO" \
    --group-name /aws/lambda/foo1
...
```

`--end` and `--start` are supported.
```

$ cwinsights query \
    --end "2020-08-15T10:00:00Z" \
    --start "2020-08-15T09:00:00Z" \
    --query-string INFO \
    --group-name /aws/lambda/foo1
```

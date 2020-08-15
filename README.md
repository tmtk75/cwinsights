# README
`cwinsights` is a utility command for CloudWatch Logs Insights for the purpose of bulky query.


## Motivation
AWS Step Functions consists of several Lambda, ECS Task, etc.
Their logs are basically stored into CloudWatch Logs and some log groups are created.
Sometimes I want to query a pattern in all such log groups.
This command supports it.


## Getting Started
`bulk` is to use the purpose. It ends synchronously.
```
$ cat > groups
/aws/lambda/foo1
/aws/lambda/foo2
/aws/lambda/foo3

$ cwinsights bulk groups \
    --before 1h \
    --query-string 'fields @message | filter @message =~ /INFO/'
...
```

Simply `query` is supported. `--group-name` gives log group name.
It's also convenient synchronously to print results.
```
$ cwinsights query \
    --before 1h \
    --query-string "INFO" \
    --group-name /aws/lambda/foo1
...
```

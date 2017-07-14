<h1 align="center">awsmon 📡  </h1>

<h5 align="center">Sends memory and disk statistics back to CloudWatch</h5>

<br/>

EC2 instances doesn't have memory and disk statistics by default. This project aims at providing a single static binary that gives you such capabilities.

> Fork of [go-aws-mon](https://github.com/a3linux/go-aws-mon/)


```
Usage: awsmon [--interval INTERVAL] [--memory] [--disk DISK] [--namespace NAMESPACE] [--aws]

Options:
  --interval INTERVAL, -i INTERVAL
                         interval between samples [default: 30s]
  --memory, -m           retrieve memory samples [default: true]
  --disk DISK, -d DISK   retrieve disk samples from disk locations [default: [/]]
  --namespace NAMESPACE, -n NAMESPACE
                         cloudwatch metric namespace [default: System/Linux]
  --aws, -a              whether or not the instance is running in aws [default: true]
  --help, -h             display this help and exit
```


### LICENSE

See `./LICENSE` (inherits from the fork).


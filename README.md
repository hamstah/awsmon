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

#### Install

You can either use `go` or fetch the binary directly from GitHub releases page:

```
go get -u github.com/cirocosta/awsmon
```

or

```sh
# change VERSION to the latest you find in `releases` section
readonly VERSION="0.1.2"
readonly URL="https://github.com/cirocosta/awsmon/releases/download/v${VERSION}/awsmon_${VERSION}_linux_amd64.tar.gz"
readonly BINARY_DESTINATION="/usr/local/bin/awsmon"

mkdir -p /tmp/awsmon
curl -o /tmp/awsmon/awsmon.tar.gz -L $URL
tar xzfv /tmp/awsmon/awsmon.tar.gz -C /tmp/awsmon
sudo mv /tmp/awsmon/awsmon $BINARY_DESTINATION
```


#### Running it while instance is alive

In order to keep the binary running through the whole life of the instance, `awsmon` can be configured as a `systemd` service with something like the following:

```
[Unit]
Description=awsmon

[Service]
User=root
ExecStart=/usr/local/bin/awsmon
Restart=always
RestartSec=3

[Install]
WantedBy=default.target
```


#### LICENSE

See `./LICENSE` (inherits from the fork).


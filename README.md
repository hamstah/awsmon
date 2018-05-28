<h1 align="center">awsmon 📡  </h1>

<h5 align="center">Sends memory and disk statistics back to CloudWatch</h5>

<br/>

EC2 instances doesn't have memory and disk statistics by default. This project aims at providing a single static binary that gives you such capabilities.

> Fork of [go-aws-mon](https://github.com/a3linux/go-aws-mon/) with static configuration, non-aws mode and continuous execution.


```
Usage: awsmon [opts...]

Options:
  --interval INTERVAL    interval between samples [default: 30s]
  --memory               retrieve memory samples [default: true]
  --load-1m              retrieve load 1m avgs [default: true]
  --load-5m              retrieve load 5m avgs
  --load-15m             retrieve load 15m avgs
  --relativize-load      makes loadavg relative to cpu count [default: true]
  --disk DISK            retrieve disk samples from disk locations [default: [/]]
  --config CONFIG        path to awsmon configuration file [default: /etc/awsmon/config.json]
  --debug                toggles debugging mode
  --aws                  whether or not to enable AWS support
  --aws-access-key AWS-ACCESS-KEY
                         aws access-key with cw putMetric caps
  --aws-secret-key AWS-SECRET-KEY
                         aws secret-key with cw putMetric caps
  --aws-asg AWS-ASG      autoscaling group that the instance is in
  --aws-instance-id AWS-INSTANCE-ID
                         id of the instance (required if wanting AWS support)
  --aws-instance-type AWS-INSTANCE-TYPE
                         type of the instance (required if wanting AWS support)
  --aws-namespace AWS-NAMESPACE
                         cloudwatch metric namespace [default: System/Linux]
  --aws-region AWS-REGION
                         region for sending cloudwatch metrics to
  --aws-aggregated-only
                         region for sending cloudwatch metrics to
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

#### Configuration

The parameters can also be statically configured via the configuration file (`--config`) that defaults to `/etc/awsmon/config.json`:

```json
{
  "interval": 30000000000,
  "memory": true,
  "disk": [
    "/"
  ],
  "debug": false,
  "aws": true,
  "aws-autoscaling-group": "",
  "aws-instance-id": "",
  "aws-instance-type": "",
  "aws-namespace": "System/Linux"
}
```

If all the instance configurations are specified, no calls are made to the EC2 metadata service. Otherwise, such information is automatically retrieved from such service.


#### Necessary permissions

The only permission needed by AWSMON is `cloudwatch:putMetricData`. 

If you're unsure of how to create an instance that has such capability, check out the `./example` directory. 

It contains [terraform](https://terraform.io) files that create an instance with the right permissions.


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


#### AWS

AWS support requires that you have already configured session support by either configuring an instance role for the EC2 instance or creating a well-formed credentials file (`~/.aws/credentials`). 

If you're unsure about whether the metrics are really being succesfully sent to CloudWatch, enable debug (append `--debug` to the configuration). This will print out the AWS client logs.


#### LICENSE

See `./LICENSE` (inherits from the fork).


#!/bin/bash

set -o errexit

readonly AWSMON_URL="https://github.com/cirocosta/awsmon/releases/download/v2.8.0/awsmon_2.8.0_linux_amd64.tar.gz"

main() {
  fetch_awsmon
}

fetch_awsmon() {
  curl -SL -o awsmon $AWSMON_URL
  tar xzf ./awsmon
  mv ./awsmon /usr/localbin
}

main

#!/usr/bin/env bash

set -e

# Setup for travis ci

# Install etcd
mkdir etcd
pushd etcd
curl -L https://github.com/coreos/etcd/releases/download/v3.1.8/etcd-v3.1.8-linux-amd64.tar.gz -o etcd-v3.1.8-linux-amd64.tar.gz
tar xzvf etcd-v3.1.8-linux-amd64.tar.gz
popd

# Install consul
mkdir consul
pushd consul
curl -L https://releases.hashicorp.com/consul/0.6.3/consul_0.6.3_linux_amd64.zip -o consul_0.6.3_linux_amd64.zip
unzip consul_0.6.3_linux_amd64.zip
popd

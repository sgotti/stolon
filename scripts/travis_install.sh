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

# Install postgreSQL 9.5, 9.6, 10
sudo /etc/init.d/postgresql stop

sudo sh -c 'echo "deb http://apt.postgresql.org/pub/repos/apt/ $(lsb_release -cs)-pgdg main 10" > /etc/apt/sources.list.d/pgdg.list'
wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | sudo apt-key add -
sudo apt-get update
sudo apt-get -y install postgresql-9.5 postgresql-9.6 postgresql-10


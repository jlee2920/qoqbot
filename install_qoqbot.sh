#!/bin/bash

if [ ! -d /var/tmp/qoqbot ]; then
    mkdir -p /var/tmp/qoqbot
else
    rm -rf /var/tmp/qoqbot/*
fi

mv /tmp/qoqbot.tgz /var/tmp/qoqbot

# Install docker
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
sudo add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
sudo apt-get update
apt-cache policy docker-ce
sudo apt-get install -y docker-ce

# Install docker-compose
sudo curl -L https://github.com/docker/compose/releases/download/1.21.2/docker-compose-$(uname -s)-$(uname -m) -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

sudo apt-get install -f

# Install fish and make fish the default shell
sudo apt-get install fish -y
sudo chsh -s `which fish`

# Upload qoqbot docker image and load it
cd /var/tmp/qoqbot
gunzip qoqbot.tgz
sudo docker load -i qoqbot.tar
cd /

# Make new directory and move docker-compose to a new directory
sudo mkdir -p /opt/qoqbot
sudo mv /tmp/docker-compose.yml /opt/qoqbot

rm -rf /var/tmp/qoqbot
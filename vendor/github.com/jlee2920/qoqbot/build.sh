#!/bin/bash

docker-compose -f docker-compose.yml up -d --build
rm -rf build
mkdir build
docker save 046294321880.dkr.ecr.us-east-1.amazonaws.com/qoqbot | gzip -c > build/qoqbot.tgz
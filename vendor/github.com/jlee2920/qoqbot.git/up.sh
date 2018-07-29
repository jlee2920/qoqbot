#!/bin/bash

docker-compose up -d db
docker-compose run -d qoqbot ./qoqbot
#!/bin/bash

docker-compose up -d db
docker-compose run -v qoqbot_data -d -T --use-aliases --rm qoqbot ./qoqbot 
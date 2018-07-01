#!/bin/bash

string="!play"

tail -n 0 -F ~/go/src/PhantomBot-2.4.0.3/logs/chat/01-07-2018.txt | \
while read LINE
do
echo "$LINE" | grep -q $string
if [ $? = 0 ]
then
request=$(echo "$LINE" | grep -o "!play.*")
./postToDiscord.sh "$request"
fi
done
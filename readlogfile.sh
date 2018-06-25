#!/bin/bash

string="!play"

tail -n 0 -F ~/Documents/PhantomBot-2.4.0.3/logs/chat/25-06-2018.txt | \
while read LINE
do
echo "$LINE" | grep -q $string
if [ $? = 0 ]
then
request=$(echo "$LINE" | grep -o "!play.*")
./postToDiscord.sh "$request"
fi
done
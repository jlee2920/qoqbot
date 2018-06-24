#!/bin/bash

string="!sr"

tail -n 0 -F ~/Documents/PhantomBot-2.4.0.3/logs/chat/23-06-2018.txt | \
while read LINE
do
echo "$LINE" | grep -q $string
if [ $? = 0 ]
then
echo -e "$string found"
fi
done
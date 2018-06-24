#!/bin/bash

# update the TOKEN and the CHANNELID, rest is optional
# you may need to connect with a websocket the first time you run the bot
#   use a library like discord.py to do so

echo "$1"
message1='{"content":'
message2='}'

message="$1"

full_message=$message1"\"$message\""$message2
echo $full_message

curl -v \
-H "Authorization: Bot TOKEN" \
-H "User-Agent: myBotThing (http://some.url, v0.1)" \
-H "Content-Type: application/json" \
-d $full_message \
https://discordapp.com/api/channels/CHANNELID/messages
#!/bin/bash

# update the TOKEN and the CHANNELID, rest is optional
# you may need to connect with a websocket the first time you run the bot
#   use a library like discord.py to do so

curl -v \
-H "Authorization: Bot TOKEN" \
-H "User-Agent: myBotThing (http://some.url, v0.1)" \
-H "Content-Type: application/json" \
-X POST \
-d '{"content":"Posting as a bot"}' \
https://discordapp.com/api/channels/CHANNELID/messages
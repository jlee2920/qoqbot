#!/bin/bash

COUNTER=1
rm -rf test.txt
while [ $COUNTER -lt 78 ]; do
STRING="ping on world $COUNTER"
ping oldschool${COUNTER}.runescape.com -c 2 | cut -d= -f 4 | sed -e '/ms/!d' -e "s/ms/$STRING/g" | head -1 >> test.txt
let COUNTER=COUNTER+1
done

echo -e "Your lowest ping is: `sort -n test.txt | head -1`"

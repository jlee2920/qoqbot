FROM golang:alpine

ADD . /go/src/qoqbot.git
WORKDIR /go/src/qoqbot.git/cmd/qoqbot
ADD waitForPG.sh .
ADD regulars.txt .
RUN go build -o qoqbot
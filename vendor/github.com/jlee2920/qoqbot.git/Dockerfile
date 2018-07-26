FROM golang:1.9.2-alpine3.6

RUN apk add --no-cache curl && \
    curl -Lo /bin/rq  https://s3-eu-west-1.amazonaws.com/record-query/record-query/x86_64-unknown-linux-musl/rq && \
    chmod +x /bin/rq

RUN mkdir -p /go/src/qoqbot
ENV GOPATH=/go
WORKDIR /go/src/qoqbot

ADD vendor /go/src/
ADD Gopkg.lock /go/src/qoqbot
RUN cd /go/src && \
  cat qoqbot/Gopkg.lock | /bin/rq -tJ 'map "projects" | spread | map "name"' | cat | tr -d '"' | xargs -I % go install %/...

ADD /cmd/qoqbot/qoqbot.go ./
ADD waitForPG.sh ./
ADD regulars.txt ./

RUN go build -o qoqbot qoqbot.go
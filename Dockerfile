# Use an official Python runtime as a parent image
FROM golang:latest

ADD . / 
WORKDIR /cmd/
RUN go build qoqbot.go
CMD ["/cmd/qoqbot"]


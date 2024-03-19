FROM golang:1.22.1 AS build-stage

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . ./

ENV CGO_ENABLED=0 

ENV GOOS=linux 

RUN go build -o /chatops-awx

FROM alpine:3.19.1 AS build-release-stage

WORKDIR /

COPY --from=build-stage /chatops-awx /chatops-awx

CMD ["/chatops-awx"]

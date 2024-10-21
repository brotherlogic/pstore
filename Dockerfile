# syntax=docker/dockerfile:1

FROM golang:1.22 AS build

WORKDIR $GOPATH/src/github.com/brotherlogic/pstore

COPY go.mod ./
COPY go.sum ./

RUN mkdir proto
COPY proto/*.go ./proto/


RUN go mod download
COPY *.go ./

RUN CGO_ENABLED=0 go build -o /pstore

##
## Deploy
##
FROM gcr.io/distroless/base-debian11

WORKDIR /

COPY --from=build /pstore /pstore

EXPOSE 8080
EXPOSE 8081

USER nonroot:nonroot

ENTRYPOINT ["/pstore"]
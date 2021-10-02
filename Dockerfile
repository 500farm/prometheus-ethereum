FROM golang:alpine AS build-stage

WORKDIR /

COPY src/* go.mod go.sum /usr/local/go/src/build/
RUN cd /usr/local/go/src/build; go build -o /usr/local/bin/ethereum_exporter .

FROM alpine

WORKDIR /usr/local/bin

COPY --from=build-stage /usr/local/bin/ethereum_exporter .

ENTRYPOINT ["/usr/local/bin/ethereum_exporter"]
EXPOSE 8577

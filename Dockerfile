FROM golang:1.17-alpine
LABEL org.opencontainers.image.source = &quot;https://github.com/renehonig/kafka2websocket&quot;

RUN apk add --update --no-cache alpine-sdk bash python3 musl-dev

# compile and install librdkafka
WORKDIR /root
RUN git clone https://github.com/edenhill/librdkafka.git
WORKDIR /root/librdkafka
# checkout v1.8.0
RUN git checkout 9ded5ee
RUN /root/librdkafka/configure
RUN make
RUN make install

# copy source files and private repo dep
COPY ./k2ws/ /go/src/k2ws/

# static build the app
WORKDIR /go/src/k2ws
RUN go mod tidy
RUN go get -d ./...
RUN go build -tags musl -ldflags '-extldflags "-static"' 

# create final image
FROM alpine

COPY --from=0 /go/src/k2ws/k2ws /usr/bin/

# RUN apk --no-cache add \
#       cyrus-sasl \
#       openssl \

ENTRYPOINT ["k2ws"]

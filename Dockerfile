FROM golang:1.21.3-alpine AS builder
WORKDIR /src
COPY . /src

RUN apk add --update --no-cache make git \
    && make chinadns


FROM scratch as chinadns
WORKDIR /0990
WORKDIR bin
COPY --from=builder /src/build/chinadns .
WORKDIR /0990/config
CMD ["../bin/chinadns","-c","chinadns.json"]

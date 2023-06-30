FROM golang:1.20.5 AS builder
WORKDIR app
COPY . .

RUN go env -w GOPROXY="https://goproxy.cn,direct"
RUN CGO_ENABLED=0 go build -o /bin/chinadns ./cmd/server/main.go

FROM scratch as chinadns
WORKDIR /0990
WORKDIR bin
COPY --from=builder /bin/chinadns .
WORKDIR /0990/config
CMD ["../bin/chinadns","-c","chinadns.json"]

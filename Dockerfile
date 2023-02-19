FROM golang:1.19 as modules

ADD go.mod go.sum /m/
RUN cd /m && go mod download

FROM golang:1.19 as builder

COPY --from=modules /go/pkg /go/pkg

RUN mkdir -p /app
COPY . /app
WORKDIR /app

RUN go build --ldflags '-extldflags "-static"' -o meow3 cmd/main.go

FROM linuxserver/ffmpeg:latest as ffmpeg

COPY --from=builder /app/meow3 meow3
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT [ "./meow3" ]
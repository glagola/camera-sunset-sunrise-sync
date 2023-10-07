FROM golang:1.21-alpine3.18 as builder

RUN apk update
RUN apk --no-cache add ca-certificates

WORKDIR /go/src

COPY go.mod go.sum ./

RUN go mod download && go mod verify

COPY . .

RUN cd ./cmd/camera-sunset-sunrise-sync && go build -o ./app .



FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/cmd/camera-sunset-sunrise-sync/app /app

CMD ["/app"]

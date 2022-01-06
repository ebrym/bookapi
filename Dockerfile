FROM golang:alpine as builder

RUN apk update && apk upgrade && \
    apk add --no-cache git

RUN mkdir /build 
WORKDIR /build 

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . . 

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bookApi *.go

FROM alpine:latest

RUN mkdir /app
WORKDIR /app

COPY --from=builder /build/bookApi .

CMD ["./github.com/ebrym/bookApi"]
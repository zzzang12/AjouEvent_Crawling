FROM golang:1.22.0-alpine as dev
WORKDIR /workspace
COPY go.mod go.sum ./
RUN go mod download -x
COPY . .
RUN apk add curl && \
    curl -sSfL https://raw.githubusercontent.com/cosmtrek/air/master/install.sh | sh -s -- -b $(go env GOPATH)/bin && \
    export PATH=$PATH:$GOROOT/bin:$GOPATH/bin:$(go env GOPATH)/bin
CMD air
EXPOSE 1323

FROM golang:1.22.0-alpine as build
WORKDIR /workspace
COPY go.mod go.sum ./
RUN go mod download -x
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o crawling .

FROM alpine:3.19.0 as prod
WORKDIR /workspace
COPY --from=build /workspace/crawling .
COPY config config

RUN apk add --no-cache tzdata && \
    cp /usr/share/zoneinfo/Asia/Seoul /etc/localtime && \
    echo "Asia/Seoul" > /etc/timezone

ENV TZ="Asia/Seoul"

CMD ./crawling
EXPOSE 1323

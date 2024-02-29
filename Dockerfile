FROM golang:1.22.0-alpine as build
WORKDIR /workspace
COPY go.mod go.sum ./
RUN go mod download -x
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o notifier .

FROM alpine:3.19.0 as run
WORKDIR /workspace
COPY --from=build /workspace/notifier .
COPY config/notifierConfigs.json .
COPY serviceAccountKey.json .
CMD ./notifier
EXPOSE 8888

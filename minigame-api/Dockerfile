FROM golang:1.22-alpine AS build

WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/minigame-api ./cmd/server

FROM alpine:3.20

WORKDIR /app
COPY --from=build /out/minigame-api /app/minigame-api
EXPOSE 8090
ENTRYPOINT ["/app/minigame-api"]

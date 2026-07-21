FROM golang:1.22-alpine AS build

WORKDIR /src
ARG GOPROXY=https://goproxy.cn,direct
ENV GOPROXY=${GOPROXY}
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/minigame-api ./cmd/server

FROM alpine:3.20

WORKDIR /app
COPY --from=build /out/minigame-api /app/minigame-api
EXPOSE 8090
ENTRYPOINT ["/app/minigame-api"]

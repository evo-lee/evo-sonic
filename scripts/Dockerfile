# syntax=docker/dockerfile:1.7

FROM golang:1.25-alpine AS builder

RUN apk --no-cache add git ca-certificates gcc g++ musl-dev

WORKDIR /go/src/github.com/evo-lee/evo-sonic

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

ARG BUILD_COMMIT
ARG BUILD_TIME
ARG SONIC_VERSION
ARG TARGETOS
ARG TARGETARCH

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=1 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
    go build -o sonic \
        -ldflags="-s -w \
            -X github.com/evo-lee/evo-sonic/consts.SonicVersion=${SONIC_VERSION} \
            -X github.com/evo-lee/evo-sonic/consts.BuildCommit=${BUILD_COMMIT} \
            -X github.com/evo-lee/evo-sonic/consts.BuildTime=${BUILD_TIME}" \
        -trimpath .

RUN mkdir -p /app/conf /app/resources && \
    cp /go/src/github.com/evo-lee/evo-sonic/sonic /app/ && \
    cp -r /go/src/github.com/evo-lee/evo-sonic/conf/. /app/conf/ && \
    cp -r /go/src/github.com/evo-lee/evo-sonic/resources/. /app/resources/ && \
    cp /go/src/github.com/evo-lee/evo-sonic/scripts/docker_init.sh /app/ && \
    chmod +x /app/docker_init.sh


FROM alpine:3.20 AS prod

RUN apk add --no-cache tzdata ca-certificates wget && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

COPY --from=builder /app/ /app/

VOLUME /sonic
EXPOSE 8080
WORKDIR /sonic

HEALTHCHECK --interval=30s --timeout=5s --start-period=20s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8080/ >/dev/null 2>&1 || exit 1

CMD ["/bin/sh", "-c", "/app/docker_init.sh && /app/sonic -config /sonic/conf/config.yaml"]

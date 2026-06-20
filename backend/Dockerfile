FROM golang:1.26-alpine AS builder

ARG GOPROXY=https://goproxy.cn,direct
ARG APK_MIRROR=mirrors.aliyun.com
ENV GOPROXY=$GOPROXY

WORKDIR /app

RUN sed -i "s/dl-cdn.alpinelinux.org/${APK_MIRROR}/g" /etc/apk/repositories && \
    apk add --no-cache build-base git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -trimpath -ldflags="-s -w" -o main ./cmd

FROM alpine:3.23

ARG APK_MIRROR=mirrors.aliyun.com

RUN sed -i "s/dl-cdn.alpinelinux.org/${APK_MIRROR}/g" /etc/apk/repositories && \
    apk --no-cache add ca-certificates tzdata

RUN cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
  echo "Asia/Shanghai" > /etc/timezone

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/config ./config

RUN mkdir -p /app/data /app/log && \
  chown -R appuser:appgroup /app

USER appuser

EXPOSE 25610

CMD ["./main"]

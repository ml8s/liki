# 多阶段构建：镜像内编译最新源码（消除预编译 bin/liki-engine 的中间步骤）
FROM golang:1.26-alpine AS build
WORKDIR /src
ENV GOPROXY=https://goproxy.cn,direct
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/liki-engine ./cmd/liki/

FROM alpine:3.21 AS certs
RUN apk add --no-cache ca-certificates
RUN adduser -D -u 1000 liki

FROM scratch
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=certs /etc/passwd /etc/passwd
COPY --from=build /out/liki-engine /usr/local/bin/liki-engine
USER liki
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/liki-engine", "--addr", ":8080"]

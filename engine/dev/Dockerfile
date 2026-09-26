# 多阶段构建：镜像只包含当前业务入口 engine-mcp。
FROM golang:1.26.6-alpine@sha256:3889b425f035be855a72fb4755265311293b6d414521f0a519d819df32222d83 AS build
WORKDIR /src
ENV GOPROXY=https://goproxy.cn,direct
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG TARGETOS=linux
ARG TARGETARCH=amd64
RUN CGO_ENABLED=0 GOOS="${TARGETOS}" GOARCH="${TARGETARCH}" \
    go build -o /out/engine-mcp ./cmd/engine-mcp/ && \
    go build -o /out/engine-rpc ./cmd/engine-rpc/

FROM alpine:3.21@sha256:48b0309ca019d89d40f670aa1bc06e426dc0931948452e8491e3d65087abc07d AS certs
RUN apk add --no-cache ca-certificates
RUN adduser -D -u 1000 liki

FROM scratch
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=certs /etc/passwd /etc/passwd
COPY --from=build /out/engine-mcp /usr/local/bin/engine-mcp
COPY --from=build /out/engine-rpc /usr/local/bin/engine-rpc
USER liki
EXPOSE 8080 8081
ENTRYPOINT ["/usr/local/bin/engine-rpc", "--addr", ":8080"]

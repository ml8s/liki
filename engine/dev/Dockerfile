# 多阶段构建：镜像含 engine-rpc（默认入口，liki-web 过渡，见 docs/RUNTIME.md）与 engine-mcp 两个入口。
FROM golang:1.26.9-alpine@sha256:cdfd4fe2da6b225d8b40c6b7a105736e548e83ff56d5d8f9394446eeb5eb84e0 AS build
WORKDIR /src
ENV GOPROXY=https://goproxy.cn,direct
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_TIME=unknown
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
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_TIME=unknown
LABEL org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${COMMIT}" \
      org.opencontainers.image.created="${BUILD_TIME}" \
      org.opencontainers.image.title="liki-engine" \
      org.opencontainers.image.description="Liki deterministic metaphysics engine"
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=certs /etc/passwd /etc/passwd
COPY --from=build /out/engine-mcp /usr/local/bin/engine-mcp
COPY --from=build /out/engine-rpc /usr/local/bin/engine-rpc
USER liki
EXPOSE 8080 8081
ENTRYPOINT ["/usr/local/bin/engine-rpc", "--addr", ":8080"]

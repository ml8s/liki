# 多阶段构建：镜像仅含 engine-mcp 运行时。
FROM golang:1.26.10-alpine@sha256:5cb67f19c5e7994dcb80c074f7cfca8aeeada45bbdca035f73df269cefdcce61 AS build
WORKDIR /src
ENV GOPROXY=https://goproxy.cn,direct
ARG VERSION=local
ARG COMMIT=unknown
ARG BUILD_TIME=unknown
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG TARGETOS=linux
ARG TARGETARCH=amd64
RUN CGO_ENABLED=0 GOOS="${TARGETOS}" GOARCH="${TARGETARCH}" \
    go build -o /out/engine-mcp ./cmd/engine-mcp/

FROM alpine:3.21@sha256:48b0309ca019d89d40f670aa1bc06e426dc0931948452e8491e3d65087abc07d AS certs
RUN apk add --no-cache ca-certificates
RUN adduser -D -u 1000 liki

FROM scratch
ARG VERSION=local
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
USER liki
EXPOSE 8081
ENTRYPOINT ["/usr/local/bin/engine-mcp", "--addr", ":8081"]

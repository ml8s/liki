# uv 基础镜像默认走 NJU 境内镜像（ghcr.io 直连不可靠）；海外构建可用
# --build-arg UV_IMAGE=ghcr.io/astral-sh/uv:0.11.20 覆盖。digest 固定不变。
ARG UV_IMAGE=ghcr.nju.edu.cn/astral-sh/uv:0.11.20
ARG UV_INDEX_URL=https://pypi.tuna.tsinghua.edu.cn/simple
# uv stage 置于最前：被 deps 引用，避免前向引用触发的循环依赖误报。
FROM ${UV_IMAGE}@sha256:eaa5f1a3305307aaf9e67fe2bbba1d85ebbb2d8a63bce23af21797bfafbe0f8b AS uv

FROM python:3.12-slim@sha256:2f17fc044b579bab302c2e8054d3a686e2cb9a83de48e70534b94cd8ebbe06a9 AS deps
ARG GIT_SHA=unknown
ARG COMMIT=unknown
ARG BUILD_TIME=unknown

COPY --from=uv /uv /usr/local/bin/uv
COPY counsel/requirements.lock /tmp/requirements.lock
ARG UV_INDEX_URL
ENV UV_INDEX_URL=$UV_INDEX_URL
RUN uv venv /opt/venv && \
    UV_HTTP_TIMEOUT=300 UV_CONCURRENT_DOWNLOADS=10 \
    uv pip install --python /opt/venv/bin/python \
      --require-hashes --requirement /tmp/requirements.lock

FROM python:3.12-slim@sha256:2f17fc044b579bab302c2e8054d3a686e2cb9a83de48e70534b94cd8ebbe06a9
ARG GIT_SHA=unknown
ARG COMMIT=unknown
ARG BUILD_TIME=unknown

WORKDIR /app

COPY --from=deps /opt/venv /opt/venv
ENV PATH="/opt/venv/bin:$PATH"

COPY counsel/app ./app
ENV LIKI_GIT_SHA="${GIT_SHA}"
LABEL org.opencontainers.image.revision="${COMMIT}" \
      org.opencontainers.image.created="${BUILD_TIME}" \
      org.opencontainers.image.title="liki-counsel" \
      org.opencontainers.image.description="Liki deterministic counsel MCP service"

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1
ENV LIKI_BIND_HOST=0.0.0.0

USER 65534:65534

EXPOSE 8086

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s \
    CMD ["python", "-c", "import urllib.request; urllib.request.urlopen('http://127.0.0.1:8086/healthz', timeout=2)"]

CMD ["sh", "-c", "exec python -m uvicorn app.counsel_mcp:app --host \"${LIKI_BIND_HOST:-0.0.0.0}\" --port 8086"]

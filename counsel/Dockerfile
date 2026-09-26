FROM python:3.12-slim@sha256:2f17fc044b579bab302c2e8054d3a686e2cb9a83de48e70534b94cd8ebbe06a9 AS deps

COPY --from=ghcr.io/astral-sh/uv:0.11.20@sha256:eaa5f1a3305307aaf9e67fe2bbba1d85ebbb2d8a63bce23af21797bfafbe0f8b /uv /usr/local/bin/uv
COPY counsel/requirements.lock /tmp/requirements.lock
RUN uv venv /opt/venv && \
    UV_HTTP_TIMEOUT=300 UV_CONCURRENT_DOWNLOADS=10 \
    uv pip install --python /opt/venv/bin/python \
      --require-hashes --requirement /tmp/requirements.lock

FROM python:3.12-slim@sha256:2f17fc044b579bab302c2e8054d3a686e2cb9a83de48e70534b94cd8ebbe06a9

WORKDIR /app

COPY --from=deps /opt/venv /opt/venv
ENV PATH="/opt/venv/bin:$PATH"

COPY counsel/app ./app

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    LIKI_MCP_URL=http://engine-mcp:8081/mcp

USER 65534:65534

EXPOSE 8086

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s \
    CMD ["python", "-c", "import urllib.request; urllib.request.urlopen('http://127.0.0.1:8086/healthz', timeout=2)"]

CMD ["python", "-m", "uvicorn", "app.counsel_mcp:app", "--host", "0.0.0.0", "--port", "8086"]

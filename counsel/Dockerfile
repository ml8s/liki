FROM python:3.12-slim

ARG LIKI_COUNSEL_SERVICE_DOMAIN=bazi

WORKDIR /app

RUN pip install --no-cache-dir "mcp" "jsonschema>=4,<5" "uvicorn"

COPY counsel/app ./app

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    LIKI_MCP_URL=http://engine-mcp:8081/mcp \
    LIKI_COUNSEL_SERVICE_DOMAIN=${LIKI_COUNSEL_SERVICE_DOMAIN}

USER 65534:65534

EXPOSE 8091

CMD ["python", "-m", "uvicorn", "app.counsel_server:app", "--host", "0.0.0.0", "--port", "8091"]

FROM python:3.12-slim

WORKDIR /app

RUN pip install --no-cache-dir "mcp" "jsonschema>=4,<5" "uvicorn"

COPY counsel/app ./app

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    LIKI_MCP_URL=http://engine-mcp:8085/mcp

USER 65534:65534

EXPOSE 8086

CMD ["python", "-m", "uvicorn", "app.counsel_server:app", "--host", "0.0.0.0", "--port", "8086"]

FROM python:3.12-slim AS runtime

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    PIP_NO_CACHE_DIR=1 \
    PIP_DISABLE_PIP_VERSION_CHECK=1 \
    PIP_INDEX_URL=https://pypi.tuna.tsinghua.edu.cn/simple

WORKDIR /app

COPY requirements.txt /app/requirements.txt
RUN python -m pip install --no-cache-dir -r /app/requirements.txt \
 && python -m pip check

COPY app /app/app
COPY app/expose.json /app/expose.json

RUN useradd --uid 10001 --no-create-home liki
USER 10001

# SQLite order store volume (paywall enabled deployments)
VOLUME ["/data"]

EXPOSE 8087
CMD ["python", "-m", "app.main"]

FROM python:3.12-slim@sha256:46cb7cc2877e60fbd5e21a9ae6115c30ace7a077b9f8772da879e4590c18c2e3

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1

WORKDIR /app

RUN pip install --no-cache-dir uv

COPY pyproject.toml README.md /app/
COPY app /app/app

RUN uv pip install --system .

EXPOSE 8000

CMD ["sh", "-c", "exec uvicorn app.main:app --host 0.0.0.0 --port ${PORT:-8000} --workers ${WEB_CONCURRENCY:-1}"]

LABEL org.opencontainers.image.source="https://github.com/kube-rca/kuberca"
LABEL org.opencontainers.image.licenses="Apache-2.0"
LABEL org.opencontainers.image.description="KubeRCA Agent"
LABEL org.opencontainers.image.documentation="https://github.com/kube-rca/kuberca/tree/main/agent"
LABEL org.opencontainers.image.vendor="KubeRCA"
LABEL org.opencontainers.image.title="KubeRCA agent"

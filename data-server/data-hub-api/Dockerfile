FROM python:3.10-slim

WORKDIR /app

# System dependencies
# ca-certificates: trust store for TLS to external Postgres (sslmode=verify-*
# with a public CA). libpq-dev: libpq with OpenSSL for source builds; the
# psycopg2/psycopg binary wheels bundle their own TLS-enabled libpq.
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc libpq-dev ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Python dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir --upgrade pip && \
    pip install --no-cache-dir -r requirements.txt

# Application code
COPY . .

EXPOSE 8100

CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8100", "--workers", "2"]

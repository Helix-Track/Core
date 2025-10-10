FROM python:3.11-alpine

# Install dependencies
RUN apk add --no-cache \
    curl \
    wget \
    bash \
    jq

# Create test user
RUN addgroup -g 1000 testuser && \
    adduser -D -u 1000 -G testuser testuser

# Install Python packages
RUN pip install --no-cache-dir \
    requests \
    pytest \
    pytest-html \
    pytest-json-report \
    colorama \
    pyyaml

# Create directories
RUN mkdir -p /tests /test-results && \
    chown -R testuser:testuser /tests /test-results

WORKDIR /tests

USER testuser

# Default command
CMD ["python3", "run_all_tests.py"]

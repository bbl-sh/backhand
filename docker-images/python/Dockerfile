FROM python:3.11-alpine

# Avoid bytecode files and ensure immediate output
ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1

# Set working directory
WORKDIR /app

# Run as non-root user (security!)
RUN addgroup -g 1001 -S appuser && \
    adduser -u 1001 -S appuser -G appuser
USER appuser

# Default command: run solution.py (will be mounted at runtime)
CMD ["python", "/app/solution.py"]

# Use Python 3.10 (recommended for PyTorch)
FROM python:3.10-slim

# Set environment to avoid interactive prompts
ENV DEBIAN_FRONTEND=noninteractive

# Install minimal dependencies
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Upgrade pip
RUN pip install --no-cache-dir --upgrade pip

# Install PyTorch CPU-only (latest stable, as of 2025)
# Official command from pytorch.org, CPU-only version
RUN pip install --no-cache-dir torch torchvision torchaudio --index-url https://download.pytorch.org/whl/cpu

# Optional: create non-root user (recommended for security in production)
RUN addgroup --gid 1001 appgroup && \
    adduser --disabled-password --gecos '' --uid 1001 --gid 1001 appuser
USER appuser

# Default command: run Python script
CMD ["python", "/app/solution.py"]

# Dockerfile.node
FROM node:24-slim

WORKDIR /app

# Create non-root user for security
RUN addgroup --gid 1001 --system appgroup && \
    adduser --uid 1001 --system appuser && \
    chown appuser:appgroup /app

USER appuser

# Run the solution (solution.js will be mounted at runtime)
CMD ["node", "/app/solution.js"]

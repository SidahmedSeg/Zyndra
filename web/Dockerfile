# Build stage
FROM oven/bun:1-alpine AS builder

WORKDIR /app

# Copy package files
COPY package.json bun.lock ./

# Install dependencies
RUN bun install --frozen-lockfile

# Copy source code
COPY . .

# Build Next.js application
RUN bun run build

# Production stage
FROM oven/bun:1-alpine

WORKDIR /app

# Copy package files
COPY package.json bun.lock ./

# Install production dependencies only
RUN bun install --frozen-lockfile --production

# Copy built application from builder
COPY --from=builder /app/.next ./.next
COPY --from=builder /app/next.config.js ./
COPY --from=builder /app/package.json ./

# Create public directory and copy if it exists (Next.js may not create it)
RUN mkdir -p ./public
COPY --from=builder /app/public ./public

# Create non-root user (use existing if GID 1000 is taken)
RUN if ! getent group 1000 > /dev/null 2>&1; then \
        addgroup -g 1000 appuser; \
    else \
        addgroup -S appuser || true; \
    fi && \
    if ! getent passwd 1000 > /dev/null 2>&1; then \
        adduser -D -u 1000 -G appuser appuser; \
    else \
        adduser -D -S -G appuser appuser || true; \
    fi

# Change ownership
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3000 || exit 1

# Start Next.js
CMD ["bun", "run", "start"]


# Build stage
FROM node:26-alpine@sha256:725aeba2364a9b16beae49e180d83bd597dbd0b15c47f1f28875c290bfd255b9 AS builder

WORKDIR /app

# Copy package files
COPY package.json package-lock.json ./

# Install dependencies
RUN npm ci

# Copy source code
COPY . .

# Build the application
RUN npm run build

# Production stage
FROM nginx:alpine@sha256:54f2a904c251d5a34adf545a72d32515a15e08418dae0266e23be2e18c66fefa

# Patch OS packages whose Alpine fix lands before the nginx base image is
# rebuilt (e.g. libxml2 CVE-2026-6732, fixed in 2.13.9-r1). Keeps the release
# Trivy gate green without waiting on an upstream nginx rebuild.
RUN apk upgrade --no-cache

# Copy built files from builder stage
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy nginx configuration (optional, for SPA routing)
RUN echo 'server { \
    listen 80; \
    server_name localhost; \
    root /usr/share/nginx/html; \
    index index.html; \
    location / { \
        try_files $uri $uri/ /index.html; \
    } \
}' > /etc/nginx/conf.d/default.conf

# Expose port 80
EXPOSE 80

# Start nginx
CMD ["nginx", "-g", "daemon off;"]

LABEL org.opencontainers.image.source="https://github.com/kube-rca/kuberca"
LABEL org.opencontainers.image.licenses="Apache-2.0"
LABEL org.opencontainers.image.description="KubeRCA Frontend"
LABEL org.opencontainers.image.documentation="https://github.com/kube-rca/kuberca/tree/main/frontend"
LABEL org.opencontainers.image.vendor="KubeRCA"
LABEL org.opencontainers.image.title="KubeRCA frontend"


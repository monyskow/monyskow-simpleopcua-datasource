# node:20-alpine pinned by digest — floating tags can pull in different minor
# Node versions across CI runs and break reproducibility.
FROM node:20-alpine@sha256:fb4cd12c85ee03686f6af5362a0b0d56d50c58a04632e6c0fb8363f609372293

WORKDIR /app

# Copy package files (lock file required for `npm ci`)
COPY package.json package-lock.json ./

# Install dependencies deterministically from the lock file
RUN npm ci --omit=dev

# Copy server code
COPY server.js ./

# Create certificates directory
RUN mkdir -p /app/certificates

# Expose OPC UA port
EXPOSE 4840

# Run the server
CMD ["node", "server.js"]

FROM node:20-alpine

WORKDIR /app

# Copy package files
COPY package.json ./

# Install dependencies
RUN npm install --production

# Copy server code
COPY server.js ./

# Create certificates directory
RUN mkdir -p /app/certificates

# Expose OPC UA port
EXPOSE 4840

# Run the server
CMD ["node", "server.js"]

# Stage 1: Dependencies
FROM node:22-alpine AS deps
WORKDIR /build

# Copy package files
COPY package.json yarn.lock ./

# Install dependencies (cached layer)
RUN yarn install --frozen-lockfile --network-timeout 100000

# Stage 2: Builder
FROM node:22-alpine AS builder
WORKDIR /build

# Copy dependencies from previous stage
COPY --from=deps /build/node_modules ./node_modules

# Copy source code
COPY . .

# Build arguments for runtime configuration
ARG VITE_API_URL=/api
ARG VITE_APP_TITLE=LeanView

# Build the application
RUN yarn build

# Stage 3: Export static files
FROM scratch AS export
COPY --from=builder /build/dist /dist

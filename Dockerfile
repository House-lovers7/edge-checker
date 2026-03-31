# Stage 1: Build Go binary
FROM golang:1.25-alpine AS go-builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ cmd/
COPY internal/ internal/

ARG VERSION=dev
ARG COMMIT=none
ARG DATE=unknown

RUN CGO_ENABLED=0 go build \
    -ldflags "-X github.com/House-lovers7/edge-checker/internal/version.Version=${VERSION} -X github.com/House-lovers7/edge-checker/internal/version.Commit=${COMMIT} -X github.com/House-lovers7/edge-checker/internal/version.Date=${DATE}" \
    -o /edge-checker ./cmd/edge-checker

# Stage 2: Build web frontend
FROM node:22-alpine AS web-builder

WORKDIR /app/web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# Stage 3: Final image
FROM alpine:3.21

RUN apk add --no-cache ca-certificates

COPY --from=go-builder /edge-checker /usr/local/bin/edge-checker
COPY --from=web-builder /app/web/dist /opt/edge-checker/web
COPY scenarios/ /opt/edge-checker/scenarios/

WORKDIR /work

ENTRYPOINT ["edge-checker"]

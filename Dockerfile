# syntax=docker/dockerfile:1

# --- Build stage ---
FROM golang:1.24-alpine AS build
WORKDIR /src

# Cache dependencies separately from the source code.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/evently .

# --- Runtime stage: static binary on distroless, running as a non-root user ---
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/evently /app/evently

USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/app/evently"]

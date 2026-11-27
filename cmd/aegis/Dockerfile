FROM node:22-alpine AS web-build

WORKDIR /src
COPY web/package.json web/package-lock.json ./web/
RUN cd web && npm ci
COPY web/ ./web/
COPY casepulse-logo/ ./casepulse-logo/
RUN cd web && npm run build

FROM golang:1.22-alpine AS go-build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=web-build /src/web/dist ./web/dist
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/aegis ./cmd/aegis

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=go-build /out/aegis /aegis
EXPOSE 8080
ENTRYPOINT ["/aegis"]

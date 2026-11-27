FROM golang:1.27-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/clio ./cmd/clio

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/clio /clio
EXPOSE 9091
ENTRYPOINT ["/clio"]

FROM golang:1.22 AS build
ARG VERSION_OVERRIDE=0.0.0
WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 VERSION=${VERSION_OVERRIDE} make build

FROM debian:bookworm-slim AS base
RUN apt-get update && apt-get install -y ca-certificates

FROM scratch
COPY --from=base /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /usr/src/app/dist/seaway /seaway

CMD ["/seaway", "controller", "--log-level=4"]

FROM golang:1.22 AS build
ARG VERSION=0.0.0
WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath --ldflags "-s -w -X ctx.sh/seaway/pkg/build.Version=${VERSION}" -o ./dist/seaway ./pkg/cmd/seaway

FROM debian:bookworm-slim AS base
RUN apt-get update && apt-get install -y ca-certificates

FROM scratch
COPY --from=base /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /usr/src/app/dist/seaway /seaway

CMD ["/seaway", "controller", "--log-level=3"]

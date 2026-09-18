FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS build
ARG TARGETOS
ARG TARGETARCH
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags='-s -w' -o /out/spire-bundle-publisher-disk ./cmd/spire-bundle-publisher-disk

FROM alpine:3.22
COPY --from=build /out/spire-bundle-publisher-disk /usr/local/bin/spire-bundle-publisher-disk
ENTRYPOINT ["/usr/local/bin/spire-bundle-publisher-disk"]

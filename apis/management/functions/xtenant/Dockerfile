# syntax=docker/dockerfile:1

ARG GO_VERSION=1

FROM --platform=${BUILDPLATFORM} golang:${GO_VERSION} AS build

WORKDIR /src/apis
ENV CGO_ENABLED=0

COPY apis/go.work ./
COPY apis/pkg/ ./pkg/
COPY apis/management/functions/xtenant/ ./management/functions/xtenant/

WORKDIR /src/apis/management/functions/xtenant

RUN --mount=type=cache,target=/go/pkg/mod go mod download

ARG TARGETOS
ARG TARGETARCH

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /function .

FROM gcr.io/distroless/static-debian12:nonroot AS image

LABEL org.opencontainers.image.source="https://github.com/mdaops/cortex"

WORKDIR /
COPY --from=build /function /function
EXPOSE 9443
USER nonroot:nonroot
ENTRYPOINT ["/function"]

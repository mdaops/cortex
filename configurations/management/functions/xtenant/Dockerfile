# syntax=docker/dockerfile:1

ARG GO_VERSION=1

FROM --platform=${BUILDPLATFORM} golang:${GO_VERSION} AS build

WORKDIR /src
ENV CGO_ENABLED=0

COPY configurations/pkg/resources/ /src/configurations/pkg/resources/
COPY configurations/management/functions/xtenant/ /src/configurations/management/functions/xtenant/

WORKDIR /src/configurations/management/functions/xtenant

RUN --mount=type=cache,target=/go/pkg/mod go mod download

ARG TARGETOS
ARG TARGETARCH

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /function .

FROM gcr.io/distroless/static-debian12:nonroot AS image
WORKDIR /
COPY --from=build /function /function
EXPOSE 9443
USER nonroot:nonroot
ENTRYPOINT ["/function"]

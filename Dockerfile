# Build the manager binary
ARG TARGETOS=linux
ARG TARGETARCH=amd64
FROM golang:1.25.5 AS builder

ARG TARGETOS=linux
ARG TARGETARCH=amd64

WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY api/ api/
COPY common/ common/
COPY controllers/ controllers/
COPY ndb_api/ ndb_api/
COPY ndb_client/ ndb_client/
COPY controller_adapters/ controller_adapters/
COPY main.go main.go

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -v -o manager main.go

# Use distroless as minimal base image to package the manager binary
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]

# Build the manager binary
FROM golang:1.26.2 AS builder

ARG TARGETOS=linux
ARG TARGETARCH=amd64

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY api/ api/
COPY common/ common/
COPY controllers/ controllers/
COPY ndb_api/ ndb_api/
COPY ndb_client/ ndb_client/
COPY controller_adapters/ controller_adapters/
COPY main.go main.go


# Build
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -v -o manager main.go

# Use distroless as minimal base image to package the manager binary
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]

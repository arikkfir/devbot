# syntax=docker/dockerfile:1

FROM golang:1.21 as builder
ARG BACKEND
ARG TARGETOS
# no default value, to default to the environment's architecture (e.g. linux/amd64)
ARG TARGETARCH

WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY api api/
COPY cmd cmd/
COPY internal internal/
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -a -o server cmd/${BACKEND}/main.go

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/server ./server
USER 65532:65532
ENTRYPOINT ["/server"]

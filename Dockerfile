ARG GO_VERSION=1.25.8
ARG ALPINE_VERSION=3.23
ARG KUSTOMIZE_VERSION=5.8.1

FROM golang:${GO_VERSION}-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /kustomize-diff .

FROM alpine:${ALPINE_VERSION}
ARG KUSTOMIZE_VERSION
RUN apk add --no-cache git curl tar && \
    curl -sL https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv${KUSTOMIZE_VERSION}/kustomize_v${KUSTOMIZE_VERSION}_linux_amd64.tar.gz \
    | tar -xz -C /usr/local/bin && \
    chmod +x /usr/local/bin/kustomize

COPY --from=builder /kustomize-diff /usr/local/bin/kustomize-diff
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]

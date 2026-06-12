# syntax=docker/dockerfile:1.7

ARG GO_VERSION=1.23
ARG DEBIAN_VERSION=bookworm
ARG NSJAIL_VERSION=3.4

# ---- Build nsjail from source ----
FROM debian:${DEBIAN_VERSION}-slim AS nsjail-builder
ARG NSJAIL_VERSION
RUN apt-get update && apt-get install -y --no-install-recommends \
        autoconf bison ca-certificates flex g++ gcc git libnl-route-3-dev \
        libprotobuf-dev libtool make pkg-config protobuf-compiler \
    && rm -rf /var/lib/apt/lists/*
RUN git clone --depth 1 --branch ${NSJAIL_VERSION} https://github.com/google/nsjail.git /src/nsjail \
    && make -C /src/nsjail \
    && install -m 0755 /src/nsjail/nsjail /usr/local/bin/nsjail

# ---- Builder / dev image (Go + linters + nsjail) ----
FROM golang:${GO_VERSION}-${DEBIAN_VERSION} AS builder
RUN DEBIAN_FRONTEND=noninteractive apt-get update && apt-get install -y --no-install-recommends \
        libnl-route-3-200 libprotobuf32 g++ default-jdk python3 nodejs iverilog r-base ocaml jq python3-matplotlib \
    && rm -rf /var/lib/apt/lists/*
COPY --from=nsjail-builder /usr/local/bin/nsjail /usr/local/bin/nsjail
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
RUN go install github.com/tsenart/vegeta/v12@latest
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/goboxd ./cmd/goboxd

# ---- Runtime image ----
FROM debian:${DEBIAN_VERSION}-slim AS runtime
RUN DEBIAN_FRONTEND=noninteractive apt-get update && apt-get install -y --no-install-recommends \
        ca-certificates g++ default-jdk libnl-route-3-200 libprotobuf32 python3 libcap2-bin nodejs iverilog r-base ocaml \
    && rm -rf /var/lib/apt/lists/*
COPY --from=nsjail-builder /usr/local/bin/nsjail /usr/local/bin/nsjail
COPY --from=builder        /out/goboxd          /usr/local/bin/goboxd
COPY languages.yaml /etc/goboxd/languages.yaml
ENV LANGUAGES_PATH=/etc/goboxd/languages.yaml

# Configure capabilities on nsjail so non-root users can run it
RUN setcap 'cap_sys_admin,cap_setuid,cap_setgid+ep' /usr/local/bin/nsjail

# Create non-root user (available for sandboxed runs if needed, but main server runs as root to allow namespace creation)
RUN groupadd -g 1000 gobox && useradd -u 1000 -g gobox -m -s /bin/bash gobox

EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/goboxd"]

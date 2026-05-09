# -----------------------------------------------------------------------------
FROM golang:1.26-alpine AS gobuilder

# Set necessary environmet variables needed for our image
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    USER=appuser \
    UID=1001

# Install git + SSL ca certificates.
# Git is required for fetching the dependencies.
# Ca-certificates is required to call HTTPS endpoints.
RUN apk update && apk add --no-cache git ca-certificates bash tzdata && update-ca-certificates

# Create an unprivileged user
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \    
    --shell "/sbin/nologin" \    
    --no-create-home \
    --uid "${UID}" \
    "${USER}"

# Create Data Directory
RUN mkdir /etc/ha-gateway && \
    chown appuser:appuser /etc/ha-gateway && \
    chmod 755 /etc/ha-gateway

# Move to working directory /build
WORKDIR /build

# Copy and download dependency using go mod
COPY go.mod .
COPY go.sum .
RUN go mod download && \
    go mod verify

# Copy the code into the container
COPY . .

# Build the binary
RUN go build -ldflags "-w -s" -o ./ha-gateway ./cmd/...

# -----------------------------------------------------------------------------
FROM scratch

# Import from builder
COPY --from=gobuilder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=gobuilder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=gobuilder /etc/passwd /etc/passwd
COPY --from=gobuilder /etc/group /etc/group

# Setup volume
COPY --from=gobuilder --chown=appuser:appuser /etc/ha-gateway /etc/ha-gateway
VOLUME /etc/ha-gateway

# Copy our static executable
COPY --from=gobuilder /build/ha-gateway /opt/ha-gateway

# Use an unprivileged user.
USER appuser:appuser

# Run the binary.
WORKDIR /opt/
ENTRYPOINT ["./ha-gateway"]
CMD ["server"]

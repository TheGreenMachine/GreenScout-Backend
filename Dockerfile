# Copyright 2020 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Use the offical golang image to create a binary.
# This is based on Debian and sets the GOPATH to /go.
# https://hub.docker.com/_/golang
FROM golang:1.22-bookworm as builder

# Create and change to the app directory.
WORKDIR /app

# Retrieve application dependencies.
# This allows the container build to reuse cached dependencies.
# Expecting to copy go.mod and if present go.sum.
COPY go.* ./
COPY vendor vendor/

# Copy local code to the container image.
COPY . .
RUN go mod tidy && go mod vendor

# Build the binary.
RUN go build -v -o gs-backend

# Use the official Debian slim image for a lean production container.
# https://hub.docker.com/_/debian
# https://docs.docker.com/develop/develop-images/multistage-build/#use-multi-stage-builds
FROM debian:bookworm-slim
RUN set -x && (type -p wget >/dev/null || (apt update && apt-get install wget -y)) && \
    mkdir -p -m 755 /etc/apt/keyrings && \
    wget -qO- https://cli.github.com/packages/githubcli-archive-keyring.gpg | \
    tee /etc/apt/keyrings/githubcli-archive-keyring.gpg > /dev/null && \
    chmod go+r /etc/apt/keyrings/githubcli-archive-keyring.gpg && \
    echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | \
    tee /etc/apt/sources.list.d/github-cli.list > /dev/null && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y \
    ca-certificates sqlite3 pip pipx python3 git gh && \
    rm -rf /var/lib/apt/lists/*

# Copy the binary to the production image from the builder stage.
COPY --from=builder /app/gs-backend /app/gs-backend
COPY scripts/*.py entrypoint.sh /app/

WORKDIR /app
RUN chmod u+x /app/*.py && chmod u+x /app/entrypoint.sh && \
    python3 -m venv . && \
    ./bin/pip install git+https://github.com/TBA-API/tba-api-client-python.git && \
    mkdir -p /app/run && \
    mkdir -p /app/conf

ENV PATH="/app/bin:$PATH"

RUN chown -R 1001:1001 /app
USER 1001:1001

# Run the web service on container startup.
ENTRYPOINT ["/app/entrypoint.sh"]

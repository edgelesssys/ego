# Copyright (c) 2024 Intel Corporation

# Declare nodejs version you want to use
ARG NODE_VERSION=20.11.1

# Use multi-stage builds to reduce final image size
FROM docker.io/library/debian AS builder

# Define arguments used across multiple stages
ARG DCAP_VERSION=DCAP_1.21
ARG NODE_VERSION

# update and install packages, nodejs
RUN DEBIAN_FRONTEND=noninteractive \
    apt-get update -yq \
    && apt-get upgrade -yq \
    && apt-get install -yq --no-install-recommends \
        build-essential \
        ca-certificates \
        curl \
        gnupg \
        git \
        zip \
        python3 \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Install nvm (Node Version Manager)
RUN curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.7/install.sh | bash

# Set NVM_DIR so we can use it in subsequent commands
ENV NVM_DIR /root/.nvm

# Install specific version of Node using nvm
# Source nvm in each RUN command to ensure it's available
RUN . "$NVM_DIR/nvm.sh" && nvm install $NODE_VERSION && nvm use $NODE_VERSION

# Set PATH to include the node and npm binaries
ENV PATH $NVM_DIR/versions/node/v$NODE_VERSION/bin:$PATH

# Clone the specific branch or tag 
RUN git clone --recurse-submodules https://github.com/intel/SGXDataCenterAttestationPrimitives.git -b ${DCAP_VERSION} --depth 1

# Build libPCKCertSelection library
WORKDIR /SGXDataCenterAttestationPrimitives/tools/PCKCertSelection/
RUN make \
    && mkdir -p ../../QuoteGeneration/pccs/lib \
    && cp ./out/libPCKCertSelection.so ../../QuoteGeneration/pccs/lib/ \
    && make clean

# Build PCCS
WORKDIR /SGXDataCenterAttestationPrimitives/QuoteGeneration/pccs/
RUN npm config set proxy $http_proxy \
    && npm config set https-proxy $https_proxy \
    && npm config set engine-strict true \
    && npm install

# Start final image build
FROM docker.io/library/debian:12-slim

ARG NODE_VERSION

# Create user and group before copying files
ARG USER=pccs
RUN useradd -M -U -r ${USER} -s /bin/false

# Copy only necessary files from builder stage
COPY --from=builder /root/.nvm/versions/node/v$NODE_VERSION/bin/node /usr/bin/node
COPY --from=builder --chown=${USER}:${USER} /SGXDataCenterAttestationPrimitives/QuoteGeneration/pccs/ /opt/intel/pccs/

# Set the working directory and switch user
WORKDIR /opt/intel/pccs/
USER ${USER}

# Define entrypoint
ENTRYPOINT ["/usr/bin/node", "pccs_server.js"]


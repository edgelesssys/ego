# Copyright (C) 2021 Edgeless Systems GmbH. All rights reserved.
# Copyright (C) 2011-2021 Intel Corporation. All rights reserved.
#
# Redistribution and use in source and binary forms, with or without
# modification, are permitted provided that the following conditions
# are met:
#
#   * Redistributions of source code must retain the above copyright
#     notice, this list of conditions and the following disclaimer.
#   * Redistributions in binary form must reproduce the above copyright
#     notice, this list of conditions and the following disclaimer in
#     the documentation and/or other materials provided with the
#     distribution.
#   * Neither the name of Intel Corporation nor the names of its
#     contributors may be used to endorse or promote products derived
#     from this software without specific prior written permission.
#
# THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
# "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
# LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
# A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
# OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
# SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
# LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
# DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
# THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
# (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
# OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

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
# EDG: apply patch that makes apikey optional
COPY apikey.patch .
RUN git apply apikey.patch
RUN npm config set proxy $http_proxy \
    && npm config set https-proxy $https_proxy \
    && npm config set engine-strict true \
    && npm install

# Start final image build
FROM docker.io/library/debian:12-slim

# EDG: install openssl for generating temporary keys
RUN apt-get update && apt-get install -y openssl && rm -rf /var/lib/apt/lists/*

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

# EDG: override entrypoint
COPY setup.sh .
ENTRYPOINT ["./setup.sh"]

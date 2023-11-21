ARG egover=1.4.1

# Use this to build an executable for Ubuntu 22.04
FROM ghcr.io/edgelesssys/ego/build-base:v${egover} AS build

# Use this to build an executable for Ubuntu 20.04
# FROM ghcr.io/edgelesssys/ego/build-base-focal:v${egover} AS build

ARG egover

# Install required packages
# These are cached in the build-base image. Don't run `apt-get update` or
# you may get other package versions and the build won't be reproducible.
RUN apt-get install -y --no-install-recommends \
  build-essential \
  ca-certificates \
  git \
  wget

# Download and install further requirements (if any)
#
# Make sure that these stay the same, e.g., don't use "latest", but fixed versions.
#
# Avoid installing packages via apt here. This may change the version of already
# installed dependencies and may influence the final binary. If not using apt isn't
# feasible, consider building a Docker image that gathers all required apt packages
# and serves as a stable base.

# Download and install EGo
# Use --force-depends to ignore SGX dependencies, which aren't required for building
RUN egodeb=ego_${egover}_amd64_ubuntu-$(grep -oP 'VERSION_ID="\K[^"]+' /etc/os-release).deb \
  && wget https://github.com/edgelesssys/ego/releases/download/v${egover}/${egodeb} \
  && dpkg -i --force-depends ${egodeb}

# Build your app
RUN git clone -b v${egover} --depth=1 https://github.com/edgelesssys/ego \
  && cd ego/samples/helloworld \
  && ego-go build -trimpath
WORKDIR ego/samples/helloworld
RUN --mount=type=secret,id=signingkey,dst=private.pem,required=true ego sign helloworld
RUN ego bundle helloworld

# Use the deploy target if you want to deploy your app as a Docker image
FROM ghcr.io/edgelesssys/ego-deploy:v${egover} AS deploy
COPY --from=build /ego/samples/helloworld/helloworld /
ENTRYPOINT ["ego", "run", "helloworld"]

# Use the export target if you just want to use Docker to build your app and then export it
FROM scratch AS export
COPY --from=build /ego/samples/helloworld/helloworld /ego/samples/helloworld/helloworld-bundle /

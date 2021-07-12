FROM ubuntu:focal-20210609 AS build

RUN apt update && apt install -y \
  build-essential=12.8ubuntu1.1 \
  git \
  gnupg \
  wget

# download, verify, and install ego
RUN wget -qO- https://download.01.org/intel-sgx/sgx_repo/ubuntu/intel-sgx-deb.key | apt-key add \
  && echo 'deb [arch=amd64] https://download.01.org/intel-sgx/sgx_repo/ubuntu focal main' >> /etc/apt/sources.list \
  && wget https://github.com/edgelesssys/ego/releases/download/v0.3.0/ego_0.3.0_amd64.deb \
  && echo 'f646dacb851a667bfb3aad7c7314670fc9467835cd54e4bcacd5620f4ca97813  ego_0.3.0_amd64.deb' | sha256sum -c \
  && apt update && apt install -y ./ego_0.3.0_amd64.deb

# build your app
RUN git clone -b v0.3.0 --depth=1 https://github.com/edgelesssys/ego \
  && cd ego/samples/helloworld \
  && ego-go build -trimpath
RUN --mount=type=secret,id=signingkey,dst=/ego/samples/helloworld/private.pem,required=true ego sign ego/samples/helloworld/helloworld

# use the deploy target if you want to deploy your app as a Docker image
FROM ghcr.io/edgelesssys/ego-deploy AS deploy
COPY --from=build /ego/samples/helloworld/helloworld /
ENTRYPOINT ["ego", "run", "helloworld"]

# use the export target if you just want to use Docker to build your app and then export it
FROM scratch AS export
COPY --from=build /ego/samples/helloworld/helloworld /

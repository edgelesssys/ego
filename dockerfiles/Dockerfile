FROM ghcr.io/edgelesssys/ego/build-base:v1.5.2 AS build

# don't run `apt-get update` because required packages are cached in build-base for reproducibility
RUN apt-get install -y --no-install-recommends \
  build-essential \
  ca-certificates \
  clang-11 \
  cmake \
  git \
  libssl-dev \
  ninja-build \
  wget

ARG erttag=v0.4.3
ARG egotag=v1.5.2
RUN wget -qO- https://go.dev/dl/go1.21.11.linux-amd64.tar.gz | tar -C /usr/local -xz \
  && git clone -b $erttag --depth=1 https://github.com/edgelesssys/edgelessrt \
  && git clone -b $egotag --depth=1 https://github.com/edgelesssys/ego \
  && mkdir ertbuild egobuild

# install ert
RUN cd edgelessrt && export SOURCE_DATE_EPOCH=$(git log -1 --pretty=%ct) && cd /ertbuild \
  && cmake -GNinja -DCMAKE_BUILD_TYPE=Release -DBUILD_TESTS=OFF /edgelessrt \
  && ninja install

# build ego
RUN cd ego && export SOURCE_DATE_EPOCH=$(git log -1 --pretty=%ct) && cd /egobuild \
  && . /opt/edgelessrt/share/openenclave/openenclaverc \
  && cmake -DCMAKE_BUILD_TYPE=Release /ego \
  && PATH=$PATH:/usr/local/go/bin make -j`nproc` install \
  && cpack -G DEB \
  && DEBNAME=$(ls ego_*_amd64.deb) \
  && mv $DEBNAME ${DEBNAME%.*}_ubuntu-22.04.deb

FROM scratch AS export
COPY --from=build /egobuild/ego_*_amd64_ubuntu-22.04.deb /

FROM ghcr.io/edgelesssys/edgelessrt-dev AS dev
LABEL description="EGo is an SDK to build confidential enclaves in Go - as simple as conventional Go programming!"
COPY --from=build /opt/ego /opt/ego
ENV PATH=${PATH}:/opt/ego/bin

FROM ghcr.io/edgelesssys/edgelessrt-deploy AS deploy
LABEL description="A runtime version of EGo to handle enclave-related tasks such as signing and running Go SGX enclaves."
COPY --from=build /opt/ego/bin/ /opt/ego/bin
COPY --from=build /opt/ego/share /opt/ego/share
ENV PATH=${PATH}:/opt/ego/bin

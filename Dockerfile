# syntax=docker/dockerfile:1.2

FROM alpine/git:latest AS pull
RUN --mount=type=secret,id=repoaccess,dst=/root/.netrc,required=true git clone --depth 1 --recurse-submodules --shallow-submodules https://github.com/edgelesssys/ego /ego

FROM ghcr.io/edgelesssys/edgelessrt-dev:nightly AS build
COPY --from=pull /ego /ego
WORKDIR /ego/build
RUN cmake -DCMAKE_BUILD_TYPE=RelWithDebInfo ..
RUN make && make install

FROM ghcr.io/edgelesssys/edgelessrt-dev:nightly as ego-dev
LABEL description="EGo is an SDK to build confidential enclaves in Go - as simple as conventional Go programming!"
COPY --from=build /opt/ego /opt/ego
ENV PATH=${PATH}:/opt/ego/bin
ENTRYPOINT ["bash"]

FROM ghcr.io/edgelesssys/edgelessrt-deploy:nightly as ego-deploy
LABEL description="A runtime version of EGo to handle enclave-related tasks such as signing and running Go SGX enclaves."
COPY --from=build /opt/ego/bin/ego /opt/ego/bin/ego
COPY --from=build /opt/ego/share /opt/ego/share
ENV PATH=${PATH}:/opt/ego/bin
ENTRYPOINT ["bash"]

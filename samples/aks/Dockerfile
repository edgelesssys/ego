# syntax=docker/dockerfile:experimental

FROM ghcr.io/edgelesssys/ego-dev AS build
RUN git clone https://github.com/edgelesssys/ego /ego
WORKDIR /ego/samples/remote_attestation
RUN ego-go build
RUN --mount=type=secret,id=signingkey,dst=/ego/samples/remote_attestation/private.pem,required=true ego sign server

FROM ghcr.io/edgelesssys/ego-deploy AS deploy
LABEL description="EGo sample image"
COPY --from=build /ego/samples/remote_attestation/server /remote_attestation/
ENV AZDCAP_DEBUG_LOG_LEVEL=error
ENTRYPOINT [ "ego", "run", "/remote_attestation/server" ]

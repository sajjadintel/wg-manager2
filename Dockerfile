FROM mullvadvpn/golang@sha256:5a8d2c3dba8153e415273bb07e2ca7e55da2b045420ae3f3a322edcbf477b8f9

# Get NFPM
ENV NFPM_VERSION 0.11.0
ENV NFPM_DOWNLOAD_URL https://github.com/goreleaser/nfpm/releases/download/v$NFPM_VERSION/nfpm_amd64.deb
ENV NFPM_DOWNLOAD_SHA256 b489c1eebf06327b6a327ec64893b32557450609b1768839a459e769bf9d824d

RUN apt-get update && apt-get install -y curl git pkg-config libipset-dev && curl -fsSL "$NFPM_DOWNLOAD_URL" -o nfpm_amd64.deb \
    && echo "$NFPM_DOWNLOAD_SHA256  nfpm_amd64.deb" | sha256sum -c - \
    && dpkg -i nfpm_amd64.deb \
    && rm nfpm_amd64.deb

ADD . /wireguard-manager
WORKDIR /wireguard-manager

CMD ["make", "nfpm"]

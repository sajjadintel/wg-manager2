FROM mullvadvpn/golang@sha256:ff860e740b0f5ea09c6d9c2b513ae6a751aecc77015dc5ea0a962b8fa981df5c

# Get NFPM
ENV NFPM_VERSION 0.11.0
ENV NFPM_DOWNLOAD_URL https://github.com/goreleaser/nfpm/releases/download/v$NFPM_VERSION/nfpm_amd64.deb
ENV NFPM_DOWNLOAD_SHA256 b489c1eebf06327b6a327ec64893b32557450609b1768839a459e769bf9d824d

RUN apt-get update && apt-get install -y curl git && curl -fsSL "$NFPM_DOWNLOAD_URL" -o nfpm_amd64.deb \
    && echo "$NFPM_DOWNLOAD_SHA256  nfpm_amd64.deb" | sha256sum -c - \
    && dpkg -i nfpm_amd64.deb \
    && rm nfpm_amd64.deb

ADD . /wireguard-manager
WORKDIR /wireguard-manager

CMD ["make", "nfpm"]

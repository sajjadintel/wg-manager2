FROM mullvadvpn/go-packager@sha256:a54c9376a54d5b1a38710a11f86fc3a093272efffb64506b337b6f5d5b265d4d

RUN apt-get update && apt-get install -y pkg-config libipset-dev

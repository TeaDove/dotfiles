ARG CODEX_VERSION=0.166.0

FROM debian:trixie-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        bash \
        bubblewrap \
        ca-certificates \
        curl \
        git \
        jq \
        less \
        nodejs \
        npm \
        openssh-client \
        procps \
        util-linux \
        wget \
    && rm -rf /var/lib/apt/lists/*

RUN case "$(dpkg --print-architecture)" in \
        amd64) checksum=cc232108374e8e2b537045f2d5eaa1229f619db314c3ba6a14d38b35dc0aa123 ;; \
        arm64) checksum=4ae22576a78223d8ccdf5f3b3660a5cbb5288416bf0ca96afc4a36d195932b94 ;; \
        *) exit 1 ;; \
    esac \
    && curl -fsSL --retry 3 \
        "https://github.com/aptible/supercronic/releases/download/v0.2.51/supercronic-linux-$(dpkg --print-architecture)" \
        -o /usr/local/bin/supercronic \
    && printf "%s  /usr/local/bin/supercronic\n" "$checksum" | sha256sum -c - \
    && chmod 755 /usr/local/bin/supercronic

RUN useradd --create-home --uid 1000 --shell /bin/bash aibot \
    && mkdir -p /workspace \
    && chown aibot:aibot /workspace

COPY --chmod=755 entrypoint.sh /usr/local/bin/entrypoint.sh

USER aibot
WORKDIR /workspace

ENV PATH="/home/aibot/.local/bin:${PATH}"
ENV CODEX_HOME=/home/aibot/.codex

ARG CODEX_VERSION
RUN mkdir -m 700 /home/aibot/.ssh \
    && npm install --prefix /home/aibot/.local --global @openai/codex@${CODEX_VERSION} \
    && codex remote-control --no-daemon --help

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]

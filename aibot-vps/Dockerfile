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
        wget \
    && rm -rf /var/lib/apt/lists/*

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

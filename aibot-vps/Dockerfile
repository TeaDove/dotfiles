FROM debian:trixie-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        bash \
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

RUN useradd --create-home --uid 1000 --shell /bin/bash claude \
    && mkdir -p /workspace \
    && chown claude:claude /workspace

COPY --chmod=755 entrypoint.sh /usr/local/bin/entrypoint.sh

USER claude
WORKDIR /workspace

ENV PATH="/home/claude/.local/bin:${PATH}"
ENV CODEX_HOME=/home/claude/.codex

ARG CODEX_VERSION=0.166.0
RUN mkdir -m 700 /home/claude/.ssh \
    && npm install --prefix /home/claude/.local --global @openai/codex@${CODEX_VERSION} \
    && codex remote-control --no-daemon --help

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]

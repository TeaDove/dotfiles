FROM debian:trixie-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        bash \
        ca-certificates \
        curl \
        git \
        jq \
        less \
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
ENV CLAUDE_CONFIG_DIR=/home/claude/.claude

RUN mkdir -m 700 /home/claude/.ssh \
    && curl -fsSL https://claude.ai/install.sh | bash

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]

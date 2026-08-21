# syntax=docker/dockerfile:1
ARG GOLANGCI_VERSION=v2.10.1

FROM golang:1.27-trixie AS go
FROM golangci/golangci-lint:${GOLANGCI_VERSION} AS golangci

FROM python:3.14-slim-trixie

ARG NODE_MAJOR=20
ARG USERNAME=dev
ARG USER_UID=1000
ARG USER_GID=1000

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y --no-install-recommends \
      ca-certificates curl git make sudo less \
      iptables ipset iproute2 dnsutils aggregate jq \
      build-essential \
    && rm -rf /var/lib/apt/lists/*

RUN curl -fsSL https://deb.nodesource.com/setup_${NODE_MAJOR}.x | bash - \
    && apt-get install -y --no-install-recommends nodejs \
    && rm -rf /var/lib/apt/lists/*

COPY --from=go /usr/local/go /usr/local/go
ENV GOPATH=/go
ENV GOMODCACHE=/go/pkg/mod
ENV GOCACHE=/home/${USERNAME}/.cache/go-build
ENV PATH=/usr/local/go/bin:/go/bin:${PATH}
ENV CLAUDE_CONFIG_DIR=/home/${USERNAME}/.claude

RUN groupadd --gid ${USER_GID} ${USERNAME} \
    && useradd --uid ${USER_UID} --gid ${USER_GID} -m -s /bin/bash ${USERNAME} \
    && echo "${USERNAME} ALL=(root) NOPASSWD: /usr/local/bin/init-firewall.sh" \
         > /etc/sudoers.d/${USERNAME} \
    && chmod 0440 /etc/sudoers.d/${USERNAME}

COPY --from=golangci /usr/bin/golangci-lint /go/bin/golangci-lint

RUN pip install --no-cache-dir pre-commit

RUN npm install -g @anthropic-ai/claude-code

COPY init-firewall.sh /usr/local/bin/init-firewall.sh
RUN chmod +x /usr/local/bin/init-firewall.sh

RUN mkdir -p /go/pkg/mod \
             /home/${USERNAME}/.cache/go-build /home/${USERNAME}/.cache/pip \
             /home/${USERNAME}/.cache/pre-commit \
             /home/${USERNAME}/.claude /home/${USERNAME}/.claude/projects \
    && chown -R ${USER_UID}:${USER_GID} /go /home/${USERNAME}

USER ${USERNAME}
WORKDIR /home/${USERNAME}

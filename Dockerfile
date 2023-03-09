FROM mcr.microsoft.com/dotnet/sdk:6.0

RUN mkdir -p /usr/share/man/man1 \
    && apt-get update \
    && curl -fsSL https://deb.nodesource.com/setup_current.x | bash - \
    && apt-get install -y --no-install-recommends \
      git \
      nodejs \
      maven \
      python3-dev \
      python3-pip \
      python3-setuptools \
      build-essential \
      liblz4-1 \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* \
    && curl -L https://go.dev/dl/go1.20.2.linux-amd64.tar.gz --output go.tar.gz && rm -rf /usr/local/go && tar -C /usr/local -xzf go.tar.gz && rm -rf go.tar.gz \
    && update-alternatives --install /usr/bin/pip pip /usr/bin/pip3 1 \
    && pip install -U pip \
    && pip install --upgrade setuptools

RUN echo "go:\n\t$(go version)" \
    && echo "node:\n\t$(node --version)" \
    && echo "npm:\n\t$(npm --version)" \
    && echo "java:\n\t$(java --version)" \
    && echo "mvn:\n\t$(mvn --version)" \
    && echo "dotnet:\n\t$(dotnet --version)" \
    && echo "python:\n\t$(python --version)" \
    && echo "pip:\n\t$(pip --version)" \
    && echo "apt list --installed:" \
    && apt list --installed

LABEL maintainer="estafette.io"

RUN npm install --unsafe-perm -g snyk \
    && snyk --version

COPY ${ESTAFETTE_GIT_NAME} /
COPY settings.xml /settings.xml

ENV ESTAFETTE_LOG_FORMAT="console"

RUN printenv

ENTRYPOINT ["/${ESTAFETTE_GIT_NAME}"]

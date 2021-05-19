FROM mcr.microsoft.com/dotnet/sdk:5.0

RUN mkdir -p /usr/share/man/man1 \
    && apt-get update \
    && apt-get install -y --no-install-recommends \
      git \
      golang-go \
      nodejs \
      npm \
      maven \
      python3-pip \
      build-essential \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* \
    && update-alternatives --install /usr/bin/pip pip /usr/bin/pip3 1 \
    && pip install --upgrade setuptools

RUN echo "go: $(go version)" >> /versions.yaml \
    && echo "node: $(node --version)" >> /versions.yaml \
    && echo "npm: $(npm --version)" >> /versions.yaml \
    # && echo "java: $(java --version)" >> /versions.yaml \
    # && echo "mvn: $(mvn --version)" >> /versions.yaml \
    && echo "dotnet: $(dotnet --version)" >> /versions.yaml \
    # && echo "python: $(python --version)" >> /versions.yaml \
    && echo "pip: $(pip --version)" >> /versions.yaml \
    && cat /versions.yaml

LABEL maintainer="estafette.io"

RUN npm install -g snyk

COPY ${ESTAFETTE_GIT_NAME} /
COPY settings.xml /settings.xml

ENV ESTAFETTE_LOG_FORMAT="console"

RUN printenv

ENTRYPOINT ["/${ESTAFETTE_GIT_NAME}"]
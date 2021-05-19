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
      # build-essential \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* \
    && update-alternatives --install /usr/bin/pip pip /usr/bin/pip3 1 \
    && pip install --upgrade setuptools

RUN export EXT_GO_VERSION="$(go version)"
RUN export EXT_NODE_VERSION="$(node --version)"
RUN export EXT_NPM_VERSION="$(npm --version)"
RUN export EXT_JAVA_VERSION="$(java --version)"
RUN export EXT_MAVEN_VERSION="$(mvn --version)"
RUN export EXT_DOTNET_VERSION="$(dotnet --version)"
RUN export EXT_PYTHON_VERSION="$(python --version)"
RUN export EXT_PIP_VERSION="$(pip --version)"

LABEL maintainer="estafette.io"

RUN npm install -g snyk

COPY ${ESTAFETTE_GIT_NAME} /
COPY settings.xml /settings.xml

ENV ESTAFETTE_LOG_FORMAT="console"

ENTRYPOINT ["/${ESTAFETTE_GIT_NAME}"]
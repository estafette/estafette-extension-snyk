FROM mcr.microsoft.com/dotnet/sdk:5.0

RUN mkdir -p /usr/share/man/man1 \
    && apt-get update \
    && apt-get install -y --no-install-recommends \
      git \
      golang-go \
      nodejs \
      npm \
      maven \
      python3-dev \
      python3-pip \
      python3-setuptools \
      build-essential \
      liblz4-1 \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* \
    && update-alternatives --install /usr/bin/pip pip /usr/bin/pip3 1 \
    && pip install -U pip \
    && pip install --upgrade setuptools

# fix vulnerabilities
RUN curl -fSL https://mirror.koddos.net/apache/commons/io/binaries/commons-io-2.9.0-bin.tar.gz | tar xvz \
    && cp commons-io-2.9.0/commons-io-2.9.0.jar /usr/share/java/commons-io.jar

RUN curl -fSL https://repo1.maven.org/maven2/com/google/guava/guava/30.1.1-jre/guava-30.1.1-jre.jar --output usr/share/java/guava.jar

RUN curl -fSL https://repo.maven.apache.org/maven2/org/apache/maven/wagon/wagon-http/2.8/wagon-http-2.8-shaded.jar --output /usr/share/java/wagon-http-shaded-3.3.1.jar

RUN echo "go:" \
    && go version \
    && echo "node:" \
    && node --version \
    && echo "npm:" \
    && npm --version \
    && echo "java:" \
    && java --version \
    && echo "mvn:" \
    && mvn --version \
    && echo "dotnet:" \
    && dotnet --version \
    && echo "python:" \
    && python --version \
    && echo "pip:" \
    && pip --version \
    && echo "apt list --installed:" \
    && apt list --installed

LABEL maintainer="estafette.io"

RUN npm install -g snyk

COPY ${ESTAFETTE_GIT_NAME} /
COPY settings.xml /settings.xml

ENV ESTAFETTE_LOG_FORMAT="console"

RUN printenv

ENTRYPOINT ["/${ESTAFETTE_GIT_NAME}"]
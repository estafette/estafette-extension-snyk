FROM mcr.microsoft.com/dotnet/sdk:5.0

# # fix vulnerabilities
# RUN apk add --update --no-cache binutils libjpeg-turbo

# # install golang
# RUN apk add --update --no-cache git go

# # install node
# RUN apk add --update --no-cache nodejs npm

# # install java
# RUN apk add --update --no-cache maven

# # install python
# RUN apk add --update --no-cache py3-pip

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
      git \
      golang-go \
      nodejs \
      npm \
      maven \
      py3-pip \
    && rm -rf /var/lib/apt/lists/*

LABEL maintainer="estafette.io"

RUN npm install -g snyk

COPY ${ESTAFETTE_GIT_NAME} /
COPY settings.xml /settings.xml

ENV ESTAFETTE_LOG_FORMAT="console"

ENTRYPOINT ["/${ESTAFETTE_GIT_NAME}"]
FROM mcr.microsoft.com/dotnet/sdk:5.0-alpine

RUN echo $HOME

# fix vulnerabilities
RUN apk add --update binutils libjpeg-turbo

# install golang
RUN apk add --no-cache git go

# install node
RUN apk add --update nodejs npm

# install java
RUN apk add --no-cache maven

LABEL maintainer="estafette.io"

RUN npm install -g snyk

COPY ${ESTAFETTE_GIT_NAME} /
COPY settings.xml /settings.xml

ENV ESTAFETTE_LOG_FORMAT="console"

ENTRYPOINT ["/${ESTAFETTE_GIT_NAME}"]
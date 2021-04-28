FROM alpine:3.13

# fix vulnerabilities
RUN apk add --update binutils libjpeg-turbo

# install golang
RUN apk add --no-cache git go

# install node
RUN apk add --update nodejs npm

# install .net core
# RUN apk add --no-cache curl bash \
#     && curl -L https://dot.net/v1/dotnet-install.sh | bash

# install java
RUN apk add --no-cache maven

LABEL maintainer="estafette.io"

RUN npm install -g snyk

COPY ${ESTAFETTE_GIT_NAME} /
COPY ca-certificates.crt /etc/ssl/certs/

ENV ESTAFETTE_LOG_FORMAT="console"

ENTRYPOINT ["/${ESTAFETTE_GIT_NAME}"]
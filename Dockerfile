FROM registry.hub.docker.com/library/golang:1.15.0-alpine3.12 AS build

COPY . /tmp/build

WORKDIR /tmp/build
RUN set -o errexit ;\
  apk add -U bash &> /dev/null; \
  ./go-executable-build.bash

FROM registry.hub.docker.com/library/alpine:3.12

COPY --from=build /tmp/build/lagrande-linux-amd64 /usr/local/bin/lagrande

ENTRYPOINT ["/usr/local/bin/lagrande"]

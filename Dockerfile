FROM alpine:latest
WORKDIR /
COPY kubesentry /usr/local/bin/kubesentry
ENTRYPOINT ["/usr/local/bin/kubesentry"]
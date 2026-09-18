FROM alpine:3.22
COPY spire-bundle-publisher-disk /usr/local/bin/spire-bundle-publisher-disk
ENTRYPOINT ["/usr/local/bin/spire-bundle-publisher-disk"]

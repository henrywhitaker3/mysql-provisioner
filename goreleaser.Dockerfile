FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY mysql-provisioner /manager
USER 65532:65532

ENTRYPOINT ["/manager"]

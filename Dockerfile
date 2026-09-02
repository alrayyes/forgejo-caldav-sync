# goreleaser cross-compiles the binary and hands it to `docker build` as
# part of the build context — this Dockerfile's job is to COPY it, not to
# `go build` it. See rules/go-releases.md: building from source here would
# pay for a `RUN go build` under QEMU emulation on every non-native
# architecture in the goreleaser matrix, instead of the fast, no-emulation
# cross-compile goreleaser's own build phase already does.
FROM gcr.io/distroless/static-debian12:nonroot@sha256:afa5c872c891853ca7fcf1f12c3edb23f7eeef36189728842dd51042ff57f7ab
ARG TARGETPLATFORM
COPY $TARGETPLATFORM/forgejo-caldav-sync /forgejo-caldav-sync
EXPOSE 8080
# Exec form, calling the binary itself with its "healthcheck" argument —
# distroless has no shell or curl for a CMD-SHELL/curl-style check to run.
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD ["/forgejo-caldav-sync", "healthcheck"]
ENTRYPOINT ["/forgejo-caldav-sync"]

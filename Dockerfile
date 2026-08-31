FROM golang:1.27.0@sha256:4013ae0f9e7994f8535c58c811f8f863fbed38b72e0d51e6592156f758d66146 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/forgejo-caldav-sync ./cmd/forgejo-caldav-sync

FROM gcr.io/distroless/static-debian12:nonroot@sha256:afa5c872c891853ca7fcf1f12c3edb23f7eeef36189728842dd51042ff57f7ab
COPY --from=build /out/forgejo-caldav-sync /forgejo-caldav-sync
EXPOSE 8080
# Exec form, calling the binary itself with its "healthcheck" argument —
# distroless has no shell or curl for a CMD-SHELL/curl-style check to run.
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD ["/forgejo-caldav-sync", "healthcheck"]
ENTRYPOINT ["/forgejo-caldav-sync"]

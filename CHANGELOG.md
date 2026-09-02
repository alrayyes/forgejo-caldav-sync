# Changelog

## [1.1.0](https://github.com/alrayyes/forgejo-caldav-sync/compare/forgejo-caldav-sync-v1.0.1...forgejo-caldav-sync-v1.1.0) (2026-09-02)


### Features

* **ci:** upload coverage to Codecov ([#7](https://github.com/alrayyes/forgejo-caldav-sync/issues/7)) ([ad2b596](https://github.com/alrayyes/forgejo-caldav-sync/commit/ad2b5961da55538f72ab7c6578cf2e4c0f9d0519)), closes [#6](https://github.com/alrayyes/forgejo-caldav-sync/issues/6)
* **cli:** adopt cobra/viper configuration layering (closes [#30](https://github.com/alrayyes/forgejo-caldav-sync/issues/30)) ([#35](https://github.com/alrayyes/forgejo-caldav-sync/issues/35)) ([a7d30c2](https://github.com/alrayyes/forgejo-caldav-sync/commit/a7d30c2e316450bf3b791a99dc59905a885df084))
* **openspec:** initialize OpenSpec, backfill specs from issue history ([#33](https://github.com/alrayyes/forgejo-caldav-sync/issues/33)) ([bf988bc](https://github.com/alrayyes/forgejo-caldav-sync/commit/bf988bca0358bad832a0bd635cde51c7cdb6a2d8))
* **release:** switch to release-please + goreleaser (closes [#29](https://github.com/alrayyes/forgejo-caldav-sync/issues/29)) ([#36](https://github.com/alrayyes/forgejo-caldav-sync/issues/36)) ([30b8bca](https://github.com/alrayyes/forgejo-caldav-sync/commit/30b8bca5e2f16627a307a53d7142c7657be9c656))
* sync forgejo issues into caldav calendar ([4f2456f](https://github.com/alrayyes/forgejo-caldav-sync/commit/4f2456fd0a2c2ed52b4c26b5a136b83fb59d2e25))
* sync forgejo issues into caldav calendar ([8731316](https://github.com/alrayyes/forgejo-caldav-sync/commit/8731316d42feeb9f20b2fbee336a6c6e4bb7c28b)), closes [#1](https://github.com/alrayyes/forgejo-caldav-sync/issues/1)


### Bug Fixes

* **api:** lint the openapi spec in git hooks, not just ci ([#23](https://github.com/alrayyes/forgejo-caldav-sync/issues/23)) ([eefcff3](https://github.com/alrayyes/forgejo-caldav-sync/commit/eefcff3e3659a6f0fc3bee7fe49dff58b7a4f64a))
* **ci:** lint pull request titles as conventional commits ([#28](https://github.com/alrayyes/forgejo-caldav-sync/issues/28)) ([f7e4c49](https://github.com/alrayyes/forgejo-caldav-sync/commit/f7e4c494c7566f8306b085f8affc2bd57a301a4c))
* **deps:** audit bun dependencies in ci ([#24](https://github.com/alrayyes/forgejo-caldav-sync/issues/24)) ([9ed5f5e](https://github.com/alrayyes/forgejo-caldav-sync/commit/9ed5f5efb39b9d42d8860bb92d0e0385095b1279))
* **deps:** Bump distroless/static-debian12 from `1b7b9f0` to `afa5c87` ([62a4842](https://github.com/alrayyes/forgejo-caldav-sync/commit/62a4842df361e8abfc757f2347df718e86109d7d))
* **deps:** Bump distroless/static-debian12 from `1b7b9f0` to `afa5c87` ([87db91c](https://github.com/alrayyes/forgejo-caldav-sync/commit/87db91c37468fc7bd0d379cf888dc162673408fa))
* **deps:** Bump github.com/stretchr/testify in the go-dependencies group ([#5](https://github.com/alrayyes/forgejo-caldav-sync/issues/5)) ([3d149d8](https://github.com/alrayyes/forgejo-caldav-sync/commit/3d149d89f043212604bda59a309f6323e493cff9))
* **deps:** Bump golang from 1.26.6 to 1.27.0 ([#18](https://github.com/alrayyes/forgejo-caldav-sync/issues/18)) ([9e656b5](https://github.com/alrayyes/forgejo-caldav-sync/commit/9e656b5bca76f96db17d911174171478bcd8fa55))
* **deps:** regenerate bun.lock with the pinned bun version ([67ada8d](https://github.com/alrayyes/forgejo-caldav-sync/commit/67ada8da1df180bde2eab57a32866286861c64c3))
* **docker:** add dockerignore to shrink the build context ([#25](https://github.com/alrayyes/forgejo-caldav-sync/issues/25)) ([c83b2f8](https://github.com/alrayyes/forgejo-caldav-sync/commit/c83b2f851f74fe3cf7ec74785f76d5fe0132c7bd))
* **go:** add go mod tidy -diff to pre-push and CI ([#22](https://github.com/alrayyes/forgejo-caldav-sync/issues/22)) ([8afc43b](https://github.com/alrayyes/forgejo-caldav-sync/commit/8afc43bd6868064bb11d87643247c79bdf590bbb))
* **go:** bump go.mod and CI to go 1.27.0 ([#21](https://github.com/alrayyes/forgejo-caldav-sync/issues/21)) ([4271c88](https://github.com/alrayyes/forgejo-caldav-sync/commit/4271c881c0c535eec2ad93972fb18caff06bceae))
* **lint:** run golangci-lint at push, not commit ([#20](https://github.com/alrayyes/forgejo-caldav-sync/issues/20)) ([e5e512f](https://github.com/alrayyes/forgejo-caldav-sync/commit/e5e512f08a516b258db4c06495df46e4e9134c22))
* **release:** parse the release PR JSON in bash, not fromJSON() ([#37](https://github.com/alrayyes/forgejo-caldav-sync/issues/37)) ([c2e15c5](https://github.com/alrayyes/forgejo-caldav-sync/commit/c2e15c5230aca7334d43ff076c0449b98ea0bd28))
* **release:** wire changelog and git plugins into semantic-release ([#16](https://github.com/alrayyes/forgejo-caldav-sync/issues/16)) ([0217ab0](https://github.com/alrayyes/forgejo-caldav-sync/commit/0217ab009262a244a47e185d56c0b7de63fb2a89)), closes [#15](https://github.com/alrayyes/forgejo-caldav-sync/issues/15)
* **tooling:** add missing prettier config ([#32](https://github.com/alrayyes/forgejo-caldav-sync/issues/32)) ([840f030](https://github.com/alrayyes/forgejo-caldav-sync/commit/840f0303b68b0e9304c5f3035720b272c550aee1))

## [1.0.1](https://github.com/alrayyes/forgejo-caldav-sync/compare/v1.0.0...v1.0.1) (2026-08-30)


### Bug Fixes

* **release:** wire changelog and git plugins into semantic-release ([#16](https://github.com/alrayyes/forgejo-caldav-sync/issues/16)) ([0217ab0](https://github.com/alrayyes/forgejo-caldav-sync/commit/0217ab009262a244a47e185d56c0b7de63fb2a89)), closes [#15](https://github.com/alrayyes/forgejo-caldav-sync/issues/15)

# SPIRE Bundle Publisher Disk

A public SPIRE BundlePublisher plugin that will atomically write the current trust bundle to a configured filesystem path. The output directory can be mounted into a container or exported read-only through NFS for agent bootstrap.

This repository was created from the [`go-repo-template`](https://github.com/faisal-memon/go-repo-template). It targets the current SPIRE Plugin SDK and produces statically linked ARM64 and AMD64 binaries for mounting into the SPIRE server container.

## Development

Run `make lint`, `make test`, and `make govulncheck` locally. The plugin should write through a temporary file and rename it into place so readers never observe a partial bundle. The SPIRE server loads the binary through `plugin_cmd`; it does not need a separate publisher container.

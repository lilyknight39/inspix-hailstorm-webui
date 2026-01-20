# inspix-hailstorm

This is a fork of https://github.com/vertesan/inspix-hailstorm. Thanks to the original author and contributors.

An all-in-one tool suite to analyze/decrypt Link Like Love Live resources and the master database, with a CLI and an experimental WebUI.

For the Chinese version, see `README.zh-CN.md`.

> Parts of the implementation are referenced from: https://github.com/AllenHeartcore/GkmasObjectManager

## Features

- Download and decrypt assets and the master database
- Database structure analysis mode for developers
- Experimental WebUI for browsing and search
- Docker workflow support

## Quick Start

You can build from source or run via container.

### Build from Source

```bash
go build .
```

After building, use `./hailstorm -h` for options. Common flags:

- No flags: download and decrypt all new assets and DB since the last run
- `--analyze`: analyze database structure for developers
- `--dbonly`: database only, skip assets
- `--web`: start WebUI (default `127.0.0.1:5001`)

### WebUI

Run from project root:

```bash
go run . --web --addr 127.0.0.1:5001
```

Or run after build:

```bash
./hailstorm --web --addr 127.0.0.1:5001
```

Open in browser: `http://127.0.0.1:5001`.

### WebUI Preview Extras

The WebUI can generate richer previews for some asset types when optional tools
are available:

- ACB audio preview: requires `vgmstream-cli` and `ffmpeg` in `PATH`.
  Preview output is cached under `cache/webui-preview/acb`.
- Unity assetbundle preview: set `HAILSTORM_ASSETRIPPER_CMD` to a command that
  exports images/videos/models into an output folder. The command receives
  `{input}` and `{output}` placeholders. Output is cached under
  `cache/webui-preview/assetbundle/<label>`.

Example (replace with your AssetRipper export command):

```bash
export HAILSTORM_ASSETRIPPER_CMD="assetripper-export --input {input} --output {output}"
```

### Docker

Image:
https://github.com/vertesan/inspix-hailstorm/pkgs/container/inspix-hailstorm

Run `run_docker.sh` to start the container.
Defaults to `dbonly`; override via docker CLI `--entrypoint` if needed.

## References & Thanks

- https://github.com/vertesan/inspix-hailstorm
- https://github.com/AllenHeartcore/GkmasObjectManager
- https://github.com/AssetRipper/AssetRipper

## License

AGPL-3.0

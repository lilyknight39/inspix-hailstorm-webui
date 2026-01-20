# inspix-hailstorm

本项目为 https://github.com/vertesan/inspix-hailstorm 的 fork，感谢原作者与贡献者。

用于解析连结喜欢爱生活资源与主数据库的一体化工具套件，包含命令行与试验性 WebUI。

英文版本请见 `README.md`。

> 部分实现参考了项目：https://github.com/AllenHeartcore/GkmasObjectManager

## 功能概览

- 资源与主数据库的下载、解密
- 数据结构分析模式（面向开发者）
- 试验性 WebUI 浏览与检索
- Docker 运行方式

## 快速开始

你可以自行从源码构建，或使用容器运行。

### 源码构建

```bash
go build .
```

构建完成后可用 `./hailstorm -h` 查看参数说明，常用参数：

- 无参数：下载并解密自上次运行以来的所有新资源与数据库
- `--analyze`：开发者分析数据库结构
- `--dbonly`：仅处理数据库，不下载资源
- `--web`：启动 WebUI（默认地址 `127.0.0.1:5001`）

### WebUI

项目根目录下启动：

```bash
go run . --web --addr 127.0.0.1:5001
```

或构建后运行：

```bash
./hailstorm --web --addr 127.0.0.1:5001
```

打开浏览器访问：`http://127.0.0.1:5001`。

### WebUI 预览增强

当可选工具可用时，WebUI 支持更丰富的资源预览能力：

- ACB 音频预览：需要 `vgmstream-cli` 与 `ffmpeg` 已加入 `PATH`。
  预览输出缓存于 `cache/webui-preview/acb`。
- Unity assetbundle 预览：通过 `HAILSTORM_ASSETRIPPER_CMD` 设置导出命令，
  命令支持 `{input}` 与 `{output}` 占位符，输出缓存于
  `cache/webui-preview/assetbundle/<label>`。

示例（请替换为你自己的 AssetRipper 导出命令）：

```bash
export HAILSTORM_ASSETRIPPER_CMD="assetripper-export --input {input} --output {output}"
```

### Docker

镜像地址：
https://github.com/vertesan/inspix-hailstorm/pkgs/container/inspix-hailstorm

运行 `run_docker.sh` 启动容器。
默认以 `dbonly` 模式执行，如需更改可通过 docker CLI 传入 `--entrypoint` 覆盖。

## 参考与致谢

- https://github.com/vertesan/inspix-hailstorm
- https://github.com/AllenHeartcore/GkmasObjectManager
- https://github.com/AssetRipper/AssetRipper

## License

AGPL-3.0

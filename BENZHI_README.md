# satellite-pass-planning-service__001

基于 Go 实现的卫星过站规划服务 Web 项目，一款后端服务，处理过站窗口的计划、状态流转、批量导出与地面站作业记录管理。
## 构建镜像

请从**仓库根目录**执行；`benzhi.Dockerfile`、`build_benzhi_docker.sh`、`BENZHI_README.md` 均固定在该目录：

```bash
./build_benzhi_docker.sh <image-name> [linux/amd64|linux/arm64]
```

## 标准命令

```bash
go build ./...     # 编译
go run .   # 启动
go test ./...      # 测试（如有）
```

## 环境

- 基础镜像: golang:1.23
- Go 模块目录: `backend`
- 依赖已在镜像构建阶段预下载，容器内离线可用。
- 容器内工作目录: `/app/backend`

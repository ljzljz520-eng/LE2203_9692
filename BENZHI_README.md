# LE2203_9692

基于 Go 实现的命令行项目，一款业务数据管理工具，完成数据登记、校验审核、流程发布与报告导出。

## Standard commands

```bash
go build ./...
go test -count=1 ./...
```

## Run

```bash
go run ./cmd/firmwarehub
```

## Docker validation

```bash
chmod +x build_benzhi_docker.sh
./build_benzhi_docker.sh my-go-task linux/arm64
./build_benzhi_docker.sh my-go-task linux/amd64
docker run -it my-go-task:latest
```

## Known initial failures

See `BUG_REPRO.md` for the exact command and output captured during packaging.

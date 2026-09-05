BINARY  := bin/nscardprice
PKG     := ./cmd/nscardprice

.PHONY: build install test integration clean

## build         编译到 bin/nscardprice
build:
	go build -o $(BINARY) $(PKG)

## install       构建并安装到 ~/go/bin（PATH 直接用 nscardprice）
install: build
	go install $(PKG)

## test          运行单元测试（不联网）
test:
	go test ./...

## integration   运行真实接口集成测试（间隔 3s，防封 IP）
integration:
	NSCARD_INTEGRATION=1 go test -tags integration ./test/ -v

## clean         删除本地二进制
clean:
	rm -f $(BINARY)
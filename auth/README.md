# auth

jwt auth

- 自定义JWT`authutil`：依赖Redis，参考：[kratos子项目:beer-shop](https://github.com/go-kratos/beer-shop)
- 标准JWT`jwtutil`：摘自[kratos子项目:beer-shop](https://github.com/go-kratos/beer-shop)

## protobuf

```shell
protoc \
		--proto_path=. \
        --proto_path=./third_party \
	    --go_out=paths=source_relative:. \
	    ./auth/*.proto
```

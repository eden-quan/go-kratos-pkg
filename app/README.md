# app

```shell
protoc \
		--proto_path=. \
	    --go_out=paths=source_relative:. \
	    ./app/*.proto
```
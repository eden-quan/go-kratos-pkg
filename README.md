# go kratos pkg

## 添加私有包

添加环境变量

```shell

# golang 环境变量
# windows系统，请在系统环境变量中添加
# 也可使用 go env -w 写入；例如 go env -w GO111MODULE=on
# 如果在环境变量中添加过了了。go env -w 会包系统变量冲突错误
export GO111MODULE=on
export GOPROXY=https://goproxy.cn,direct
export GOPRIVATE=some.gitlab.cn
export GOPATH="/path/to/go" # windows请在环境变量Path中添加目录$GOPATH/bin
export GOBINPATH="$GOPATH/bin" # windows请在环境变量Path中添加目录$GOPATH/bin
export PATH="$PATH:$GOBINPATH" # windows请在环境变量Path中添加目录$GOPATH/bin

```

## 将下载代码方式由https改为ssh


```shell
# 配置git-ssh
git config --global url."ssh://git@gitlab.some.cn/".insteadOf "https://some.gitlab.cn/"
```

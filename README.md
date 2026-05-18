# gf-wrapper

gf-wrapper 是基于 [GoFrame](https://goframe.org/) [gf CLI](https://github.com/gogf/gf) 的增强版命令行工具。

内嵌官方 GoFrame CLI 命令树：

```go
github.com/gogf/gf/cmd/gf/v2/gfcmd
```

本项目只在执行官方命令前增加一步：按 GoFrame 官方 `gcfg.AdapterFile` 流程找到配置文件，读取配置内容，渲染显式 secret/env 占位符，再通过 `AdapterFile.SetContent` 注入回配置适配器，最后继续执行官方 `gfcmd`。

没有新增 CLI 开关。是否渲染完全由配置文件里的占位符决定；没有占位符时，应尽量等同官方 `gf` 行为。

## 安装

```bash
go install github.com/Invincib-e/gf-wrapper/cmd/gf/v2@vX.Y.Z
```

Dockerfile 示例：

```dockerfile
ARG GO_GF_VERSION=v2.10.2
RUN go install github.com/Invincib-e/gf-wrapper/cmd/gf/v2@${GO_GF_VERSION}
```

## 使用

最终命令仍然叫 `gf`，用法仍然是官方 GoFrame CLI 的用法：

```bash
gf gen dao
```

## 配置示例

```yaml
gfcli:
  gen:
    dao:
      - link: "mysql:${env:MYSQL_APP_USER}:${env:MYSQL_APP_PASSWORD}@tcp(${env:MYSQL_ADDR})/${env:MYSQL_APP_DATABASE}?charset=utf8mb4&parseTime=True&loc=Local"
        tables: ""
        group: "${env:MYSQL_APP_DATABASE}"
        removePrefix: ""
        daoPath: "internal/dao"
        doPath: "internal/model/do"
        entityPath: "internal/model/entity"
```

## 支持语法

只处理显式 source 语法：

```text
${env:KEY}
${file:/path/to/file}
${docker-secret:name}
```

### `${env:KEY}`

读取环境变量：

```go
os.LookupEnv(KEY)
```

`KEY` 必须符合常见环境变量命名格式：

```text
[A-Za-z_][A-Za-z0-9_]*
```

环境变量值会原样返回，不会二次展开。

### `${file:/path/to/file}`

读取指定文件内容，支持绝对路径和相对路径。读取结果会执行：

```go
strings.TrimRight(value, "\r\n")
```

文件内容不会二次渲染。

### `${docker-secret:name}`

默认读取：

```text
/run/secrets/name
```

等价于：

```text
${file:/run/secrets/name}
```

示例：

```yaml
password: "${docker-secret:mysql_password}"
```

`name` 只允许：

```text
[A-Za-z0-9_.-]+
```

且不能包含 `/`、`\` 或 `..`。

## 不支持语法

```text
$KEY
${KEY}
```

这些写法不会被处理。这样可以避免误伤密码或普通字符串里的 `$`，例如：

```text
abc$def
```

## 转义

以下写法会输出字面量 `${env:KEY}`，不会读取环境变量：

```text
$${env:KEY}
\${env:KEY}
```

## Kubernetes Secret

暂不提供专门的 `${k8s-secret:...}`。Kubernetes Secret 推荐两种方式：

1. 通过 env 注入后使用 `${env:KEY}`
2. 通过 volume 挂载后使用 `${file:/path/to/key}`

Kubernetes Secret 的挂载路径由 Pod spec 决定，没有统一固定路径。

## Vault

暂未实现 Vault resolver。后续可以通过新增 resolver 扩展。

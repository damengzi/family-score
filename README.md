# 家庭德育积分系统

基于 `openspec/tech_design.md` 的本机部署版 Go 应用，后端 HTTP 路由基于 `Gin` 框架，数据存储使用本地 `SQLite`，日志使用 `zap` 输出到控制台和本地文件。

## 启动

```bash
make run
```

仅构建可执行文件：

```bash
make build
```

也可以直接使用 Go 命令运行：

```bash
go run .
```

固定管理员账号：

```text
用户名：admin
密码：654321
```

默认访问：

```text
http://127.0.0.1:8080
```

## 本地数据

默认数据目录：

```text
~/.family-score/
```

包含：

- `family-score.db`：SQLite 数据库
- `backups/`：本地备份
- `logs/app.log`：zap 本地日志文件
- `files/`：预留证据文件目录
- `exports/`：预留导出目录

可通过环境变量覆盖：

```bash
FAMILY_SCORE_DATA_DIR=/path/to/data FAMILY_SCORE_ADDR=127.0.0.1:8080 go run .
```

## 角色权限

- `ADMIN` 管理员：固定账号 `admin / 654321`，负责创建/注销家长和孩子账号，也可使用业务功能。
- `PARENT` 家长：只能管理自己名下的孩子，可添加孩子、加扣分、审核任务、配置任务和奖励。
- `CHILD` 孩子：只能查看自己分值和积分明细，可自主完成日常任务、申请兑换物品，不能加扣分、创建/删除任务或奖励。

任务审核规则：孩子可自主选择任务完成，但必须由家长/管理员审核后才加分；同一孩子当天任务审核加分总额不超过 `15` 分。

## 优雅关闭

服务启动后会写入 PID 文件：

```text
~/.family-score/app.pid
```

前台运行时可用 `Ctrl+C` 触发优雅关闭。后台运行时可执行：

```bash
kill -TERM $(cat ~/.family-score/app.pid)
```

关闭过程会等待 HTTP 请求结束、关闭 SQLite 连接并删除 PID 文件。

如果端口被旧进程占用，可先查看：

```bash
lsof -nP -iTCP:8080 -sTCP:LISTEN
```

## 文档

设计文档和技术方案已统一放到 `openspec/`：

- `openspec/design.md`
- `openspec/tech_design.md`
- `openspec/design.docx`

# AGENTS.md

本文档用于指导后续 AI Agent / 维护者在本项目中进行需求分析、代码修改、验证和交付。

## 项目概览

`family-score` 是一个本机部署的家庭德育积分系统，适用于家长和孩子在家庭场景中管理基础分、兑换分、星星、家庭小队分、任务、奖励兑换和用户权限。

核心技术栈：

- 后端：`Go` + `Gin`
- 数据库：本机 `SQLite`
- 日志：`zap`
- 前端：原生 `HTML/CSS/JavaScript`
- 部署方式：个人电脑本机运行
- 数据目录：默认 `~/.family-score/`

## 常用命令

```bash
make build
```

构建可执行文件到：

```text
bin/family-score
```

```bash
make run
```

构建并运行服务，默认地址：

```text
http://127.0.0.1:8080
```

覆盖启动参数：

```bash
make run ADDR=127.0.0.1:9090
make run DATA_DIR=/tmp/family-score-data
```

直接运行：

```bash
go run .
```

验证：

```bash
go build ./...
make build
find web -name '*.js' -print0 | xargs -0 -n1 node --check
```

## 目录结构说明

```text
main.go                         应用入口、优雅启停、PID 文件
internal/consts/                业务常量
internal/controller/            HTTP 入参出参处理
internal/logger/                zap 日志初始化
internal/protocol/              请求、响应、领域模型结构体
internal/repository/            SQLite 初始化、迁移、默认数据
internal/router/                Gin 路由和静态资源分发
internal/service/               业务逻辑
pkg/utils/                      与业务无关的通用工具方法
web/                            前端静态资源
web/pages/                      前端功能页
openspec/                       需求设计和技术方案文档
```

## 当前核心功能

### 用户和角色

系统当前角色：

- `ADMIN`：管理员
- `PARENT`：家长
- `CHILD`：孩子

功能边界：

- 管理员：管理用户、孩子、任务、奖励、本机备份等全量能力。
- 家长：管理自己可操作的孩子，进行加扣分、任务审核、奖励审核、任务/奖励配置。
- 孩子：查看自己的分值、积分明细，提交任务，申请兑换奖励。

### 登录和安全

- Cookie 会话名：`fs_session`
- 登录态默认有效期：`12h`
- 当天密码错误次数超过 `5` 次后锁定账号 `15` 分钟。
- 忘记密码通过图片验证码重置密码。
- 登录后可在个人主页修改自己的密码，接口为 `POST /api/profile/password`，必须校验旧密码。
- 登录页面不得展示管理员账号和密码。

注意：当前管理员账号首次补齐时使用固定默认密码；已有 `admin` 账号启动时只校正角色和启用状态，不再重置已有密码。涉及文件：

```text
internal/consts/business_consts.go
internal/service/family_score.go
```

后续如增强安全性，可继续考虑禁止或限制忘记密码重置管理员账号、增加 CSRF/Origin 校验。

### 孩子管理

- 家长或管理员可新增孩子。
- 新增孩子时可同步创建孩子登录账号。
- 可修改孩子姓名、年龄、性别、归属家长。
- 可删除孩子；删除时会清理关联积分、任务、兑换、周/月记录，并注销绑定的孩子账号。

### 积分体系

账户维度：

- `baseScore`：基准德育分
- `bonusScore`：超额兑换分
- `starCount`：星星
- `teamScore`：家庭小队分

记录类型：

- `ADD`：加分
- `DEDUCT`：扣分
- `REPAIR`：惩罚/修复
- `TEAM`：家庭小队分
- `STAR`：星星
- `EXCHANGE`：兑换

状态等级：

- `GREEN`
- `BLUE`
- `YELLOW`
- `ORANGE`
- `RED`
- `DEEP_REPAIR`

### 任务和奖励

- 每日任务由任务模板生成。
- 孩子提交任务后，家长或管理员审核。
- 同一孩子当天任务审核加分总额不能超过 `15` 分。
- 孩子可申请兑换奖励，家长或管理员审核。

### 本机备份

- `本机备份` Tab 仅管理员可见。
- 创建备份和查看备份记录后端也必须校验管理员权限。
- 备份目录默认位于：

```text
~/.family-score/backups/
```

## 前端约定

前端使用原生 JS，不引入框架。页面按功能拆分在 `web/pages/` 下。

重要文件：

```text
web/core.js              全局状态、API 请求、角色判断、主题切换
web/favicon.svg          浏览器标签页图标，家庭成员与星星图形
web/pages/auth.js        初始化、登录、忘记密码
web/pages/layout.js      主布局、Tab、事件绑定
web/pages/profile.js     个人主页、主题/语言偏好、修改密码
web/pages/children.js    孩子管理
web/pages/users.js       用户管理
web/pages/tasks.js       今日任务
web/pages/overview.js    今日概览、角色化首页聚合
web/pages/auditCenter.js 待审核中心，聚合任务和兑换审核
web/pages/growthReport.js 成长报告，基于现有前端数据生成摘要
web/pages/wishlist.js    本地奖励愿望单工具，使用 localStorage
web/pages/rewards.js     奖励兑换
web/pages/score.js       加扣分
web/pages/scoreGuide.js  分值说明
web/pages/system.js      本机备份
web/styles.css           全局样式和角色主题
```

视觉主题要求：

- 页面背景根据角色切换：管理员、家长、孩子、未登录。
- 按钮背景不要随角色切换，保持统一主按钮风格。
- 可用角色强调色修饰页面背景、说明标签、指标卡片，但不要影响主要操作按钮的一致性。
- 个人主页支持主题偏好：`system`、`light`、`dark`；该偏好保存在浏览器 `localStorage`，不写入数据库。
- 个人主页支持语言偏好，目前覆盖个人主页、主导航、主 Hero、角色/状态/任务/奖励枚举和主要操作反馈；后续如要全站完整国际化，应继续逐页抽取长篇业务说明文案，不要在业务逻辑中硬编码语言分支。
- 个人主页入口位于右上角用户区域，不放在主 Tab 列表；个人主页中的创建时间、登录有效期按 `YYYY年MM月DD日 HH时mm分ss秒` 展示。
- 浏览器图标使用 `web/favicon.svg`，路由同时提供 `/favicon.svg` 和 `/favicon.ico`。

### 近期前端体验增强记录

已完成的体验增强以原生 JS/CSS 为主，除个人主页改密新增当前用户安全接口外，不改数据库结构：

- 新增 `今日概览`：作为登录后的默认入口，按角色展示今日任务、分值状态、奖励小货架、最新成长记录和待办摘要。
- 新增 `待审核中心`：家长/管理员可集中处理所有可访问孩子的待确认任务和待审核兑换；仍复用原有审核接口和后端权限校验。
- 新增 `成长报告`：基于当前孩子的任务、积分记录、兑换申请和愿望单，在前端生成今日简报、近 7 天趋势和本月关键词。
- 新增 `奖励愿望单`：使用浏览器 `localStorage` 按 `childId` 保存，不跨设备同步；奖励删除后前端会过滤失效项。
- 新增 `个人主页`：全角色可从右上角用户入口进入，查看当前账号、角色、绑定孩子、会话有效期、使用小结和安全建议；可设置主题、语言并修改自己的密码。
- 任务页已从表格升级为卡片，并增加“今日实践建议”；任务配置页增加“任务建议库”，点击只填入表单，不自动提交。
- 奖励页升级为奖励商店，增加健康兑换原则；奖励配置页增加“奖励配置建议”，点击只填入表单，不自动提交。
- 加扣分页增加“行为记录助手”和常用场景模板；点击模板只填入现有加扣分表单，仍需用户手动提交并接受后端校验。
- 分值说明页增加家庭场景案例、修复闭环和亲子沟通话术，用于指导实际家庭应用。

维护注意：上述快捷模板、主题/语言偏好和愿望单属于前端辅助能力，不应绕过后端权限、积分规则、任务审核上限或兑换审核规则。若后续要让偏好或愿望单跨设备同步，需要新增后端模型、迁移和权限校验。个人主页改密是后端能力，必须始终只允许修改当前登录用户且校验旧密码。

## 后端约定

### 分层

- `controller`：只处理 HTTP 请求/响应、状态码、JSON 解析。
- `service`：业务规则、权限校验、事务编排。
- `repository`：SQLite 连接、迁移、默认数据。
- `protocol`：请求/响应结构体和领域模型。
- `consts`：业务枚举和固定配置。
- `pkg/utils`：与业务无关的通用工具方法。

### 数据库

- 数据库操作统一使用 SQL，不引入 GORM 链式 ORM。
- 表结构变化必须在 `internal/repository/sqlite_store.go` 的 `Migrate()` 中补充。
- 新库要能直接创建最新完整表结构。
- 老库要通过 `addColumnIfMissing` 或显式迁移补齐字段。
- 涉及默认数据修正时，使用幂等 SQL，确保重复启动安全。

### 权限

新增接口时必须明确角色边界：

- 管理员全量权限。
- 家长通常只允许操作自己有权访问的孩子。
- 孩子只能访问绑定的 `child_id`。
- 未知角色默认拒绝。

涉及孩子数据时优先复用：

```go
s.CanAccessChild(ctx, sess, childID)
```

### 代码大小

单个代码文件原则上不超过 `1000` 行。新增逻辑应按功能拆文件，避免继续膨胀：

- 登录安全：`auth_security.go`
- 用户管理：`user_management.go`
- 孩子管理：`child_management.go`
- 积分主流程：`family_score.go`
- 业务规则：`score_rules.go`

## 修改代码时的注意事项

1. 不要把仅为单测替身服务的函数变量注入点加进业务代码。
2. 不要把本地数据库、日志、PID、构建产物提交到 Git。
3. 不要删除 `.codebuddy/` 目录。
4. 修改数据库结构后必须验证新库和旧库迁移路径。
5. 修改前端后至少执行：

```bash
find web -name '*.js' -print0 | xargs -0 -n1 node --check
```

6. 修改后端后至少执行：

```bash
go build ./...
make build
```

7. 如果涉及权限，必须同时检查前端可见性和后端接口权限，不能只隐藏按钮。

## 当前已知可优先改进点

后续 Agent 如继续优化，优先考虑：

1. 忘记密码功能禁止或限制重置管理员账号。
2. 家长权限边界进一步收紧，尤其是孩子归属转移和兑换审核。
3. 任务审核改为事务内原子处理，避免并发重复加分。
4. 兑换审核时二次校验奖励状态、孩子基准分和账户余额。
5. 给 POST/PATCH/DELETE 增加基础 CSRF/Origin 校验。
6. 孩子端“今日成长进度”、家长端“待审核提醒”、管理员端“家庭总览”已具备前端入口，后续可继续增强后端聚合接口和更完整统计口径。

## 提交前检查清单

提交或推送前请确认：

```bash
git status --short
go build ./...
make build
find web -name '*.js' -print0 | xargs -0 -n1 node --check
```

并检查：

- 是否误提交 `bin/`、数据库、日志、PID 文件。
- 是否新增表字段但未写迁移。
- 是否只做了前端隐藏而没有后端权限校验。
- 是否破坏了 `ADMIN/PARENT/CHILD` 三类角色入口。

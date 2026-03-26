# [后端基础指令]

## 当前服务边界
- gateway-service: API 网关，不持有独立数据库。
- core-service: 核心业务服务，负责 journals、wifi_space、journal_view_event、recommendation_log、journal_embedding。
- profile-service: 账户与身份服务，负责 users、auth_identity。

## 当前环境基线
- Java: 工作区固定使用 server/.vfox/sdks/java 下的 Java 21。
- Maven: 工作区固定使用 server/.vfox/sdks/maven/bin/mvn.cmd。
- Postgres: 本地 Docker，用户名 postgres，密码 root。
- Redis: 本地 Docker，默认 localhost:6379。

## 当前云端运行基线
- 正式域名入口: sunyufei5.art，由 Nginx 接管 80/443 并反向代理到 gateway-service:8080。
- 云服务器公网主入口当前为 Nginx，不应继续新增任何直接裸露到公网的后端端口。
- PostgreSQL 在云端只允许绑定服务器本机 127.0.0.1:15432，供 SSH 转发和 DBeaver 使用，禁止直接暴露公网 5432。
- 数据库查看优先级: 先容器内 psql，其次 SSH 转发 + DBeaver，最后才考虑额外开放端口；默认不走最后一种。

## 当前端口约定
- gateway-service: 8080
- core-service: 8081
- profile-service: 8082
- gateway debug: 5005
- core debug: 5006
- profile debug: 5007

## 当前数据库拆分
- resonance_profile: 账户与身份数据。
- resonance_core: 日记、空间、推荐、AI 嵌入数据。
- 严禁跨库外键；core 中的 user_id 仅保存 UUID 值，不写 REFERENCES users(id)。

## 开发优先级
1. 先保证本地 Docker 基础设施可用。
2. 先实现 profile/core 的最小业务接口，再接 gateway 转发。
3. AI 分析仍视为外部 API，不单独拆库。

## 修改后端时的默认动作
1. 保持三服务边界不被随意打穿。
2. 新增表或字段时，优先更新 sql/profile 或 sql/core 下的版本化 SQL。
3. 涉及启动或调试链路时，同时检查 .vscode/tasks.json 与 .vscode/launch.json 是否仍然匹配。
4. 更新云端后端时，只重建真正改过的服务，避免每次全量重建全部容器。
5. 后端更新完成后，至少同时验证 docker logs、actuator/health 和一个真实业务接口，禁止只看容器状态或健康检查就算完成。
6. 如果修改涉及 PostgreSQL 结构、索引、数据修复或端口映射，更新后必须补一轮 psql 或 DBeaver 实查，不能只靠接口猜测。

## 修改前端时的默认动作
1. 小程序前端更新不通过云服务器 docker compose 发布，前端交付路径固定为本地修改 client -> npm run build:weapp -> 微信开发者工具编译/上传。
2. 前端更新后，至少再走一遍“进入记事本 -> 保存记录 -> 获取回响 -> 刷新最近记录”的最小体验链路。
3. 首页和核心页面要尽量保持产品态，不要把接口调试控件、原始 JSON 和环境切换控件长期留在正式体验页。

## 代码与注释规则
1. 只要是实际业务功能代码，就要补齐中英日三语注释或说明。
2. 优先给这些位置加三语注释：Controller 接口、Service 核心流程、Entity 关键字段、推荐算法逻辑、SQL 关键字段和约束。
3. 注释不要写成模板废话；要直接解释业务意图、边界条件、降级行为和副作用。
4. 代码和注释整体风格采用 Linus Benedict Torvalds 风格：直接、具体、可验证，少空话，少包装，少为了“优雅”而绕远路。
5. 代码优先简单直白，先把数据流、控制流和错误路径写清楚，再考虑抽象层次。
6. 功能模块交付时必须附带测试用例或至少一条可执行验证路径，保证人可以验证接口、状态流转和核心结果。

## 哪些内容属于 instructions，哪些属于 skills
1. 进入 instructions 的内容，应该是长期稳定、跨任务复用且默认必须遵守的规则，例如服务边界、端口暴露策略、注释规则、验证底线、前后端发布路径。
2. 进入 skills 的内容，应该是可重复执行的流程，例如腾讯云部署、Nginx 接证书、DBeaver 走 SSH 连库、后端单服务更新、小程序体验版上传。
3. 如果一段内容描述的是“必须一直遵守什么”，写进 instructions；如果它描述的是“这件事具体怎么一步步做”，写进 skills。

## 阶段完成后的默认回看动作
1. 完成一个有明确边界的阶段后，主动检查这次变更是否需要同步更新 instructions、skills 和 MEMORY.md。
2. 如果新增的是长期规则、稳定约束或默认开发基线，写进 instructions，而不是只留在聊天记录里。
3. 如果沉淀出可重复执行的操作流程，写进 skills，而不是只靠人记忆。
4. 如果项目现状、线上拓扑、发布路径、已完成能力或已验证结论发生变化，更新 MEMORY.md 或 AGENTS.md，避免仓库知识与实际状态脱节。
# 后端最小联调验证文档

> 配套测试基线说明见 [docs/TESTING.md](docs/TESTING.md)。这份文档只负责把链路跑通，不替代自动化测试。

## 目标

这份文档只做一件事：确认 profile-service、core-service、gateway-service 组成的最小链路已经可用，前端可以开始接入。

## 前置条件

1. 本地 Docker 中的 PostgreSQL 和 Redis 已启动。
2. 三个服务已启动：
   - gateway-service: 8080
   - core-service: 8081
   - profile-service: 8082
3. 通过 gateway 访问接口，统一使用 8080 端口。

## 验证步骤

### 1. mock 登录，拿到 userId

请求：

```http
POST http://127.0.0.1:8080/api/profile/users/mock-login
Content-Type: application/json

{
  "displayName": "Sun",
  "locale": "zh-CN",
  "provider": "wechat_miniapp"
}
```

预期：

1. 返回 200。
2. 响应里包含 `id`。
3. 把这个 `id` 记为 `userId`，后面继续用。

### 2. 创建一条日记

请求：

```http
POST http://127.0.0.1:8080/api/core/journals
Content-Type: application/json

{
  "userId": "${userId}",
  "title": "礼拜一",
  "content": "今天先把最小闭环跑通。",
  "originalWeekday": 1
}
```

预期：

1. 返回 200。
2. 响应里包含日记 `id`。
3. 响应里的 `originalWeekday` 为 1。

### 3. 取一条推荐结果

请求：

```http
POST http://127.0.0.1:8080/api/core/resonance/recommend
Content-Type: application/json

{
  "userId": "${userId}",
  "hashedBssid": "demo-space-home",
  "ssidName": "Home WiFi",
  "weekday": 1
}
```

预期：

1. 返回 200。
2. `found` 为 `true`。
3. `recommendationType` 为 `SPACE_RESONANCE` 或 `WEEKDAY_RECALL`。
4. 响应中的 `journal` 不为空。

### 4. 查最近日记，确认前面写入成功

请求：

```http
GET http://127.0.0.1:8080/api/core/journals/users/${userId}
```

预期：

1. 返回 200。
2. 列表至少有 1 条。
3. 第一条应包含刚才写入的标题或内容。

## 失败时先看哪里

1. 接口 404：先看 gateway 路由配置是否仍然指向 8081 / 8082。
2. 接口 500：先看 core-service 或 profile-service 对应日志。
3. 推荐结果为空：先确认这个 `userId` 是否真的写过日记。
4. 本机联调正常但小程序真机失败：先排查请求域名、手机网络和 localhost 不可达问题。

## 别再重复踩的坑

1. 不要只看健康检查。`/actuator/health` 返回 `UP` 只能证明服务活着，不能证明业务链路真的通了。至少再打一次 `mock-login` 或写日记接口。
2. 不要只看 `docker compose ps`。容器 `Up` 不等于网关路由对、不等于数据库配置对、不等于前端可用。
3. 每次改完环境变量先复读一遍。像 `GATEWAY_PORT=808080` 这种错误，靠眼睛看 3 秒比靠日志追 30 分钟更值。
4. 上 Nginx 时先 `nginx -t` 再 reload。配置没过就先停，别拿错误配置去覆盖已经能工作的入口。

## 每次更新后的最小验证

1. 后端更新后，先看对应服务日志，再打 `actuator/health`。
2. 后端更新后，至少再打一个真实业务接口，不要只看容器和健康检查。
3. 前端更新后，重新构建 `client/dist`，然后在微信开发者工具或真机里至少走一遍“进入记事本 -> 保存记录 -> 获取回响 -> 刷新最近记录”。
4. 如果更新涉及数据库结构，除了接口验证，再补一轮 DBeaver 或 `psql` 查库验证。
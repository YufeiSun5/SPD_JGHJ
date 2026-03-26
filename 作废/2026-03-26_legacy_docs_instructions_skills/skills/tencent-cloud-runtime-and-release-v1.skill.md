# [技能] 腾讯云运行维护与小程序发布流程

## 适合写成 skill 的内容

- 云服务器初始化后的重复操作，例如同步代码、重建单个服务、检查日志、跑健康检查和真实业务接口。
- Nginx + HTTPS 的日常维护流程，例如改配置、nginx -t、reload、验证域名健康接口。
- PostgreSQL 通过 SSH 转发或 DBeaver 连接的具体步骤。
- 微信小程序前端的本地 build、开发者工具编译、上传体验版、真机回归验证流程。

## 云端服务拓扑基线

- Nginx: 80/443，对外正式入口。
- gateway-service: 8080，由 Nginx 反向代理访问。
- core-service: 8081，仅容器内网络使用。
- profile-service: 8082，仅容器内网络使用。
- PostgreSQL: 容器 5432，宿主机仅绑定 127.0.0.1:15432。
- Redis: 容器 6379，仅容器内网络使用。

## 后端更新流程

1. 先把本地修改同步到服务器 `/opt/20k-days-resonance`。
2. 登录服务器，进入 `/opt/20k-days-resonance/server/deploy`。
3. 只重建真正改过的服务，例如：

```bash
docker compose --env-file .env -f docker-compose.cloud.yml up -d --build core-service
```

4. 看该服务日志：

```bash
docker compose -f docker-compose.cloud.yml logs --tail=200 core-service
```

5. 打健康检查和真实业务接口：

```bash
curl http://127.0.0.1:8080/actuator/health
curl http://127.0.0.1:8080/api/profile/users/mock-login \
  -H 'Content-Type: application/json' \
  -d '{"displayName":"Cloud","locale":"zh-CN","provider":"wechat_miniapp"}'
```

6. 如果涉及数据库结构，再补一轮 DBeaver 或 psql 验证。

## 前端更新流程

1. 本地修改 `client`。
2. 本地打包微信小程序：

```bash
cd client
npm run build:weapp
```

3. 在微信开发者工具重新编译或重新导入 `client/dist`。
4. 至少走一遍体验链路：进入记事本 -> 保存记录 -> 获取回响 -> 刷新最近记录。
5. 需要发体验版时，在微信开发者工具点击上传，再去公众平台生成新的体验版。

## DBeaver / SSH 转发流程

1. 服务器侧必须先确认 PostgreSQL 监听 `127.0.0.1:15432`。
2. Windows 本地如 DBeaver 内置 SSH Tunnel 不稳定，直接使用系统 ssh.exe 建转发：

```powershell
& 'C:\Windows\System32\OpenSSH\ssh.exe' -N -L 15432:127.0.0.1:15432 ubuntu@49.232.169.142
```

3. DBeaver 主页面参数：
   - Host: 127.0.0.1
   - Port: 15432
   - Username: postgres
   - Password: server/deploy/.env 里的 POSTGRES_PASSWORD
   - Database: resonance_core 或 resonance_profile

## 高频排障信号

- `docker compose ps` 正常不等于业务正常，必须再打真实业务接口。
- `SCRAM-based authentication, but no password was provided`：SSH 已通，但数据库主页面没填 PostgreSQL 密码。
- `channel ... connect failed: Connection refused`：SSH 已到服务器，但服务器本机 15432 没监听，先查 compose 文件和 postgres 端口映射。
- DBeaver `测试隧道配置` 成功只代表 SSH 通道可用，不代表数据库连接成功。
- 如果命令看起来“卡住”，先判断它是否其实在等待输入；不要把省略号或提示符当命令敲进去。
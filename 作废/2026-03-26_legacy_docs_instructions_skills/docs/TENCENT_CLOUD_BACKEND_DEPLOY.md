# 腾讯云后端部署说明 / Tencent Cloud Backend Deployment / Tencent Cloud バックエンド配置手順

## 1. 目标 / Goal / 目的

- CN: 这份文档只解决第一阶段问题：把当前 backend 最小闭环部署到腾讯云服务器上，并让微信开发者工具先打通云端后端。
- EN: This document solves the first-stage goal only: deploy the current minimal backend loop to a Tencent Cloud server and let WeChat DevTools hit the cloud backend first.
- JP: この文書は第一段階だけを扱います。現在の最小バックエンド閉ループを Tencent Cloud に配置し、まず WeChat DevTools からクラウド側バックエンドへ到達させます。

- CN: 这里先不处理正式域名、Nginx、HTTPS 证书和小程序后台合法域名配置。那是第二阶段。第一阶段先把服务跑上云，先验证公网 IP 可达。
- EN: This document does not handle the formal domain, Nginx, HTTPS certificate, or Mini Program backend domain registration yet. That is phase two. Phase one is to get the services running in the cloud and validate public-IP reachability first.
- JP: ここでは正式ドメイン、Nginx、HTTPS 証明書、小程序后台の合法ドメイン設定はまだ扱いません。それは第二段階です。第一段階ではまずサービスをクラウドで動かし、公網 IP 到達性を確認します。

## 2. 当前架构 / Current Layout / 現在の構成

1. CN: `gateway-service` 对外开放端口，作为唯一公网入口。
   EN: `gateway-service` exposes the public port and acts as the only public entry.
   JP: `gateway-service` が外向けポートを公開し、唯一の公網入口になります。
2. CN: `core-service` 与 `profile-service` 仅在 Docker 内部网络通信。
   EN: `core-service` and `profile-service` communicate only on the internal Docker network.
   JP: `core-service` と `profile-service` は Docker 内部ネットワークだけで通信します。
3. CN: PostgreSQL 负责 `resonance_profile` 与 `resonance_core` 两个数据库。
   EN: PostgreSQL hosts both `resonance_profile` and `resonance_core` databases.
   JP: PostgreSQL が `resonance_profile` と `resonance_core` の二つの DB を保持します。
4. CN: Redis 负责当前最小共振链路所需缓存能力。
   EN: Redis provides the cache capability required by the current minimal resonance flow.
   JP: Redis は現在の最小共鳴フローで必要なキャッシュ機能を提供します。

## 3. 服务器前置条件 / Server Prerequisites / サーバ前提条件

1. CN: 操作系统：Ubuntu Server 24.04 LTS。
   EN: Operating system: Ubuntu Server 24.04 LTS.
   JP: OS: Ubuntu Server 24.04 LTS。
2. CN: 服务器可 SSH 登录。
   EN: The server must be reachable over SSH.
   JP: サーバへ SSH ログインできる必要があります。
3. CN: 安全组至少放行 `22` 与 `8080`。如果第二阶段接入 Nginx/HTTPS，再放行 `80` 与 `443`。
   EN: The security group must allow at least `22` and `8080`. In phase two, when Nginx/HTTPS is added, also allow `80` and `443`.
   JP: セキュリティグループでは最低 `22` と `8080` を許可します。第二段階で Nginx/HTTPS を入れるなら `80` と `443` も開けます。

## 4. 安装 Docker / Install Docker / Docker インストール

```bash
sudo apt-get update
sudo apt-get install -y ca-certificates curl gnupg
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
sudo chmod a+r /etc/apt/keyrings/docker.gpg
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  $(. /etc/os-release && echo $VERSION_CODENAME) stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt-get update
sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
sudo usermod -aG docker $USER
```

- CN: 执行完 `usermod` 后重新登录一次，让当前用户可直接执行 Docker。
- EN: Re-login after `usermod` so the current user can run Docker directly.
- JP: `usermod` 実行後、一度再ログインして現在ユーザーで Docker を直接実行できるようにします。

## 5. 上传项目 / Upload Project / プロジェクト転送

1. CN: 把整个仓库上传到服务器，例如 `/opt/20k-days-resonance`。
   EN: Upload the whole repository to the server, for example `/opt/20k-days-resonance`.
   JP: リポジトリ全体をサーバへ転送します。例: `/opt/20k-days-resonance`。
2. CN: 后续命令都假设你在仓库根目录执行。
   EN: The following commands assume you are running from the repository root.
   JP: 以下のコマンドはリポジトリルートで実行する前提です。

## 6. 准备环境变量 / Prepare Environment Variables / 環境変数準備

```bash
cd /opt/20k-days-resonance/server/deploy
cp .env.cloud.example .env
```

- CN: 至少把 `.env` 里的 `POSTGRES_PASSWORD` 换掉，别继续用默认值。
- EN: At minimum, replace `POSTGRES_PASSWORD` in `.env`; do not keep the default value.
- JP: 最低でも `.env` の `POSTGRES_PASSWORD` は変更してください。デフォルト値のまま使わないでください。

## 7. 启动服务 / Start Services / サービス起動

```bash
cd /opt/20k-days-resonance/server/deploy
docker compose --env-file .env -f docker-compose.cloud.yml up -d --build
```

- CN: 第一次启动会比较慢，因为要拉镜像、编译 jar、初始化 PostgreSQL 数据目录。
- EN: The first startup will be slower because it needs to pull images, build jars, and initialize PostgreSQL storage.
- JP: 初回起動は遅くなります。イメージ取得、jar ビルド、PostgreSQL データ初期化が走るためです。

## 8. 检查状态 / Verify Status / 状態確認

```bash
docker compose -f docker-compose.cloud.yml ps
docker compose -f docker-compose.cloud.yml logs -f gateway-service
```

健康检查：

```bash
curl http://127.0.0.1:8080/actuator/health
curl http://127.0.0.1:8080/api/profile/users/mock-login \
  -H 'Content-Type: application/json' \
  -d '{"displayName":"Cloud","locale":"zh-CN","provider":"wechat_miniapp"}'
```

- CN: 如果服务器本机能拿到 200 和 JSON，就说明公网联调只差安全组与开发者工具请求地址切换。
- EN: If the server itself returns 200 and JSON, public integration is down to security group rules and switching the DevTools request target.
- JP: サーバ自身で 200 と JSON が返るなら、公網連携で残るのはセキュリティグループ設定と DevTools 側の接続先切替だけです。

## 9. 微信开发者工具先打云端 / Hit Cloud Backend from DevTools / DevTools からクラウド接続

1. CN: 先确保腾讯云安全组放行 `8080`。
   EN: First ensure the Tencent Cloud security group allows `8080`.
   JP: まず Tencent Cloud のセキュリティグループで `8080` を開けてください。
2. CN: 在微信开发者工具里，把联调页面的 API 地址改成 `http://49.232.169.142:8080`。
   EN: In WeChat DevTools, change the workbench API base URL to `http://49.232.169.142:8080`.
   JP: WeChat DevTools でワークベンチの API 基底 URL を `http://49.232.169.142:8080` に変更します。
3. CN: 保持开发者工具里的“开发环境不校验请求域名、TLS 版本及 HTTPS 证书”开启。
   EN: Keep the DevTools option “Do not verify request domain, TLS version, and HTTPS certificate in development” enabled.
   JP: DevTools の「開発環境でリクエストドメイン、TLS バージョン、HTTPS 証明書を検証しない」を有効のままにします。
4. CN: 再按页面上的四个按钮：`mock-login -> 写日记 -> 取推荐 -> 看最近日记`。
   EN: Then press the four page actions again: `mock-login -> create journal -> get recommendation -> load recent journals`.
   JP: その後、ページ上の四つの操作を再実行します: `mock-login -> 日記作成 -> 推薦取得 -> 最近日記取得`。

## 10. 第二阶段 / Phase Two / 第二段階

- CN: 第一阶段验证通过后，不要继续把 Spring 网关直接裸露在公网端口上。第二阶段改成 `Nginx :80/:443 -> gateway-service :8080`，这才是后面给小程序正式接域名的正常结构。
- EN: After phase one passes, stop exposing the Spring gateway directly on the public edge. Phase two should become `Nginx :80/:443 -> gateway-service :8080`, which is the normal structure for a formal Mini Program domain.
- JP: 第一段階が通ったら、Spring Gateway をそのまま公網エッジに晒し続けないでください。第二段階では `Nginx :80/:443 -> gateway-service :8080` に切り替えるのが正式ドメイン運用の通常構成です。

### 10.1 安装 Nginx / Install Nginx / Nginx インストール

```bash
sudo apt-get update
sudo apt-get install -y nginx
sudo systemctl enable nginx
sudo systemctl start nginx
```

- CN: 安装完先别乱改别的，把 Nginx 跑起来再配站点。
- EN: Do not overcomplicate it after installation; start Nginx first, then configure the site.
- JP: インストール後はいきなり余計な変更をせず、まず Nginx を起動してからサイト設定に入ってください。

### 10.2 复制证书 / Copy Certificate / 証明書コピー

```bash
sudo mkdir -p /etc/nginx/ssl/sunyufei5.art
sudo cp /home/ubuntu/ssl/sunyufei5.art/sunyufei5.art_nginx/sunyufei5.art_bundle.crt /etc/nginx/ssl/sunyufei5.art/
sudo cp /home/ubuntu/ssl/sunyufei5.art/sunyufei5.art_nginx/sunyufei5.art.key /etc/nginx/ssl/sunyufei5.art/
sudo chmod 600 /etc/nginx/ssl/sunyufei5.art/sunyufei5.art.key
sudo chmod 644 /etc/nginx/ssl/sunyufei5.art/sunyufei5.art_bundle.crt
```

- CN: 证书和私钥先放到 Nginx 自己的目录，别每次都从 home 目录读。
- EN: Put the certificate and key into an Nginx-owned directory instead of reading from the home folder forever.
- JP: 証明書と秘密鍵は home 配下のまま使わず、Nginx 用ディレクトリへ移してください。

### 10.3 配置站点 / Configure Site / サイト設定

- CN: 仓库里已经提供模板文件 `server/deploy/nginx/sunyufei5.art.conf`，你可以直接照抄到服务器。
- EN: The repository now includes a template at `server/deploy/nginx/sunyufei5.art.conf`, which you can copy directly to the server.
- JP: リポジトリに `server/deploy/nginx/sunyufei5.art.conf` テンプレートを追加してあります。そのままサーバへ反映できます。

- CN: 如果云服务器上的仓库还没同步到这个文件，就别卡在那里，直接用下面的 heredoc 在服务器上现场写出来。
- EN: If the server-side repository has not been updated with this file yet, do not get stuck there; create it in place on the server with the heredoc below.
- JP: もしサーバ側リポジトリにこのファイルがまだ入っていないなら、そこで止まらず、下の heredoc でその場で作成してください。

```bash
sudo mkdir -p /opt/20k-days-resonance/server/deploy/nginx
sudo tee /opt/20k-days-resonance/server/deploy/nginx/sunyufei5.art.conf > /dev/null <<'EOF'
server {
   listen 80;
   listen [::]:80;
   server_name sunyufei5.art www.sunyufei5.art;

   return 301 https://sunyufei5.art$request_uri;
}

server {
   listen 443 ssl http2;
   listen [::]:443 ssl http2;
   server_name sunyufei5.art www.sunyufei5.art;

   ssl_certificate /etc/nginx/ssl/sunyufei5.art/sunyufei5.art_bundle.crt;
   ssl_certificate_key /etc/nginx/ssl/sunyufei5.art/sunyufei5.art.key;
   ssl_session_timeout 10m;
   ssl_session_cache shared:SSL:10m;
   ssl_protocols TLSv1.2 TLSv1.3;
   ssl_prefer_server_ciphers off;

   client_max_body_size 10m;

   location / {
      proxy_pass http://127.0.0.1:8080;
      proxy_http_version 1.1;
      proxy_set_header Host $host;
      proxy_set_header X-Real-IP $remote_addr;
      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
      proxy_set_header X-Forwarded-Proto $scheme;
      proxy_set_header Upgrade $http_upgrade;
      proxy_set_header Connection "upgrade";
      proxy_read_timeout 60s;
   }
}
EOF
```

```bash
sudo cp /opt/20k-days-resonance/server/deploy/nginx/sunyufei5.art.conf /etc/nginx/sites-available/sunyufei5.art
sudo ln -sf /etc/nginx/sites-available/sunyufei5.art /etc/nginx/sites-enabled/sunyufei5.art
sudo rm -f /etc/nginx/sites-enabled/default
sudo nginx -t
sudo systemctl reload nginx
```

- CN: `nginx -t` 不通过就先别 reload。别拿坏配置去撞生产入口。
- EN: If `nginx -t` fails, do not reload. Do not push a broken config onto the public entry.
- JP: `nginx -t` が失敗したら reload しないでください。壊れた設定をそのまま公開入口へ当てるのはダメです。

### 10.4 验证 HTTPS / Verify HTTPS / HTTPS 検証

```bash
curl -I http://sunyufei5.art
curl https://sunyufei5.art/actuator/health
curl https://sunyufei5.art/api/profile/users/mock-login \
  -H 'Content-Type: application/json' \
  -d '{"displayName":"Cloud","locale":"zh-CN","provider":"wechat_miniapp"}'
```

- CN: 第一个请求应该返回 301 或 308 跳转到 HTTPS。后两个请求应该返回健康 JSON 和用户 JSON。
- EN: The first request should return a 301 or 308 redirect to HTTPS. The next two should return health JSON and user JSON.
- JP: 最初のリクエストは HTTPS への 301 または 308 リダイレクトになるべきです。後ろ二つは health JSON と user JSON を返すべきです。

### 10.5 小程序联调切换 / Switch Mini Program Debug Target / Mini Program 接続先切替

1. CN: 把开发者工具里的 API 地址从 `http://49.232.169.142:8080` 改成 `https://sunyufei5.art`。
   EN: Change the DevTools API base URL from `http://49.232.169.142:8080` to `https://sunyufei5.art`.
   JP: DevTools の API 基底 URL を `http://49.232.169.142:8080` から `https://sunyufei5.art` に変更します。
2. CN: 开发阶段可以先继续保留“不校验合法域名”开关，等你把小程序后台服务器域名配置完再关掉。
   EN: During development you can keep the relaxed domain check enabled until the Mini Program backend domain is fully registered.
   JP: 開発中は Mini Program 側の正式ドメイン設定が完了するまで、緩いドメイン検証を一時的に維持して構いません。
3. CN: 再跑一次四步烟雾链路：`mock-login -> 写日记 -> 取推荐 -> 看最近日记`。
   EN: Run the same four-step smoke path again: `mock-login -> create journal -> get recommendation -> load recent journals`.
   JP: 同じ四段階スモークをもう一度実行します: `mock-login -> 日記作成 -> 推薦取得 -> 最近日記取得`。

## 11. 用 DBeaver 通过 SSH 连数据库 / Connect Database via DBeaver Over SSH / DBeaver で SSH 経由接続

- CN: 不要把 PostgreSQL 直接暴露到公网。当前部署把数据库只绑定到服务器本机 `127.0.0.1:15432`，然后让 DBeaver 走 SSH Tunnel。
- EN: Do not expose PostgreSQL directly to the public internet. The current deployment binds the database only to server-local `127.0.0.1:15432`, and DBeaver should connect through an SSH tunnel.
- JP: PostgreSQL を公網へ直接公開しないでください。現在の配置では DB をサーバローカル `127.0.0.1:15432` のみにバインドし、DBeaver は SSH Tunnel 経由で接続します。

### 11.1 实际服务器端口结构 / Actual Server Port Layout / 実サーバのポート構成

1. CN: `80/443` 由 Nginx 接管，对外提供正式 HTTPS 入口。
   EN: `80/443` are owned by Nginx and provide the formal HTTPS entry.
   JP: `80/443` は Nginx が受け持ち、正式な HTTPS 入口になります。
2. CN: `8080` 当前仍映射到 `gateway-service`，用于过渡期排查和内部验证，后续应该收口并关闭公网暴露。
   EN: `8080` is still mapped to `gateway-service` for transition-period troubleshooting and internal verification, but should be tightened later and closed to the public internet.
   JP: `8080` は移行期の切り分けと内部確認のためにまだ `gateway-service` へマップされていますが、後で絞り込み、公網公開を閉じるべきです。
3. CN: PostgreSQL 不对公网暴露，只绑定到服务器本机 `127.0.0.1:15432`，专门给 SSH 转发和 DBeaver 用。
   EN: PostgreSQL is not exposed publicly. It binds only to server-local `127.0.0.1:15432` for SSH forwarding and DBeaver access.
   JP: PostgreSQL は公網へ公開せず、SSH 転送と DBeaver 接続のためにサーバローカル `127.0.0.1:15432` のみにバインドします。

### 11.2 重启 PostgreSQL 容器 / Restart PostgreSQL Container / PostgreSQL 再起動

```bash
cd /opt/20k-days-resonance/server/deploy
docker compose --env-file .env -f docker-compose.cloud.yml up -d postgres
docker compose -f docker-compose.cloud.yml ps postgres
```

- CN: 正常情况下会看到 `127.0.0.1:15432->5432/tcp`，这表示数据库只对服务器本机开放，不对公网开放。
- EN: In the normal case you should see `127.0.0.1:15432->5432/tcp`, which means the database is exposed only on the server localhost and not on the public internet.
- JP: 正常なら `127.0.0.1:15432->5432/tcp` が見えます。これは DB がサーバローカルにだけ公開され、公網には出ていないことを意味します。

### 11.3 DBeaver 连接参数 / DBeaver Connection Fields / DBeaver 接続項目

1. CN: 在 PostgreSQL 连接窗口顶部，把 `URL` 切换成 `主机`。
   EN: In the PostgreSQL connection dialog, switch from `URL` to `Host` mode.
   JP: PostgreSQL 接続画面の上部で `URL` ではなく `Host` モードへ切り替えます。
2. CN: 主机填 `127.0.0.1`，端口填 `15432`。
   EN: Set Host to `127.0.0.1` and Port to `15432`.
   JP: Host は `127.0.0.1`、Port は `15432` にします。
3. CN: 数据库先填 `resonance_core` 或 `resonance_profile`，用户名填 `postgres`，密码填 `server/deploy/.env` 里的 `POSTGRES_PASSWORD`。
   EN: Set Database to `resonance_core` or `resonance_profile`, Username to `postgres`, and Password to the `POSTGRES_PASSWORD` from `server/deploy/.env`.
   JP: Database は `resonance_core` または `resonance_profile`、Username は `postgres`、Password は `server/deploy/.env` の `POSTGRES_PASSWORD` を入れます。
4. CN: 点击顶部 `SSH, SSL, ...`，打开 `SSH` 页签并勾选 `Use SSH Tunnel`。
   EN: Click `SSH, SSL, ...` at the top, open the `SSH` tab, and enable `Use SSH Tunnel`.
   JP: 上部の `SSH, SSL, ...` を押して `SSH` タブを開き、`Use SSH Tunnel` を有効にします。
5. CN: SSH 主机填 `49.232.169.142`，端口 `22`，用户 `ubuntu`，认证方式按你当前服务器登录方式选择密码或私钥。
   EN: Set SSH Host to `49.232.169.142`, Port to `22`, User to `ubuntu`, and choose password or private-key authentication according to how you currently log into the server.
   JP: SSH Host は `49.232.169.142`、Port は `22`、User は `ubuntu` にし、認証は現在のサーバログイン方法に合わせてパスワードまたは秘密鍵を選びます。
6. CN: 先点 `测试连接`，通过后再点 `完成`。
   EN: Click `Test Connection` first, and only click `Finish` after it passes.
   JP: 先に `Test Connection` を押し、通ってから `Finish` を押してください。

### 11.4 真实排障记录 / Real Debug Notes / 実際の切り分けメモ

1. CN: `测试隧道配置` 成功，不等于数据库连接成功。它只说明 DBeaver 能通过 SSH 打到服务器 `22` 端口。
   EN: A successful `Test Tunnel Configuration` does not mean the database connection works. It only proves DBeaver can reach the server `22` port over SSH.
   JP: `Test Tunnel Configuration` が成功しても、DB 接続成功を意味しません。DBeaver が SSH でサーバ `22` へ到達できることしか証明しません。
2. CN: 如果报 `The server requested SCRAM-based authentication, but no password was provided.`，说明 SSH 已通，但数据库主页面没有填 PostgreSQL 密码。
   EN: If you see `The server requested SCRAM-based authentication, but no password was provided.`, SSH is already working, but the PostgreSQL password is missing on the main connection page.
   JP: `The server requested SCRAM-based authentication, but no password was provided.` が出る場合、SSH は通っていますが、メイン接続画面で PostgreSQL パスワードが未入力です。
3. CN: 如果系统 `ssh -N -L 15432:127.0.0.1:15432 ubuntu@49.232.169.142` 报 `channel ... connect failed: Connection refused`，说明服务器本机 `15432` 没有监听，先查 `docker-compose.cloud.yml` 是否真的加了端口映射。
   EN: If system `ssh -N -L 15432:127.0.0.1:15432 ubuntu@49.232.169.142` fails with `channel ... connect failed: Connection refused`, then nothing is listening on server-local `15432`; check whether `docker-compose.cloud.yml` actually contains the port mapping first.
   JP: システム `ssh -N -L 15432:127.0.0.1:15432 ubuntu@49.232.169.142` が `channel ... connect failed: Connection refused` を返すなら、サーバローカル `15432` で待ち受けていません。まず `docker-compose.cloud.yml` に本当にポートマッピングが入っているか確認してください。
4. CN: DBeaver 内置 SSH Tunnel 可能出现“隧道测试通，但正式连接超时”的情况。遇到这种不稳定现象，优先改用系统自带 `ssh.exe` 或 `ssh -L` 做端口转发，再让 DBeaver 只连 `127.0.0.1:15432`。
   EN: DBeaver's built-in SSH tunnel can be flaky: the tunnel test may pass while the actual connection times out. In that case, prefer the system `ssh.exe` or `ssh -L` port forward, and let DBeaver connect only to `127.0.0.1:15432`.
   JP: DBeaver 内蔵 SSH Tunnel は不安定なことがあります。トンネル試験は通るのに本接続がタイムアウトする場合は、システムの `ssh.exe` または `ssh -L` に切り替え、DBeaver は `127.0.0.1:15432` だけへ接続させてください。

### 11.5 Windows 本地 SSH 端口转发 / Windows Local SSH Port Forward / Windows ローカル SSH 転送

```powershell
& 'C:\Windows\System32\OpenSSH\ssh.exe' -N -L 15432:127.0.0.1:15432 ubuntu@49.232.169.142
```

- CN: 这个窗口保持挂起是正常的，表示端口转发还活着。不要手欠把它关掉。
- EN: It is normal for this window to stay occupied. That means the port forward is alive. Do not close it casually.
- JP: このウィンドウが占有されたままなのは正常です。ポート転送が生きている状態です。むやみに閉じないでください。

## 12. 更新后端 / Update Backend / バックエンド更新

### 12.1 后端更新原则 / Backend Update Rule / バックエンド更新原則

1. CN: 只更新你改过的服务，别每次一把梭把所有容器都重建一遍。
   EN: Update only the services you actually changed. Do not rebuild everything blindly every time.
   JP: 実際に変更したサービスだけ更新してください。毎回すべてのコンテナを闇雲に作り直さないでください。
2. CN: 更新完必须先看日志，再打健康检查和真实业务接口。
   EN: After each update, inspect logs first, then hit both health and a real business endpoint.
   JP: 更新後はまずログを見て、その後に health と実際の業務 API を叩いてください。

### 12.2 同步代码到云服务器 / Sync Code To Server / サーバへコード反映

1. CN: 先把本地改动同步到服务器 `/opt/20k-days-resonance`。可以用 `scp`、WinSCP，或者你自己稳定的传输方式。
   EN: Sync your local changes to `/opt/20k-days-resonance` on the server first. Use `scp`, WinSCP, or whatever transfer method you can execute reliably.
   JP: まずローカル変更をサーバの `/opt/20k-days-resonance` へ反映します。`scp`、WinSCP、または自分が安定して使える転送方法を使ってください。
2. CN: 同步后先在服务器上 `grep` 或 `cat` 确认关键文件已经真变了，再重启容器。
   EN: After syncing, verify the key files with `grep` or `cat` on the server before restarting containers.
   JP: 反映後はコンテナ再起動前に、サーバ上で `grep` や `cat` を使って主要ファイルが本当に変わったか確認してください。

### 12.3 更新单个后端服务 / Update One Backend Service / 単一バックエンドサービス更新

示例：只更新 `core-service`

```bash
cd /opt/20k-days-resonance/server/deploy
docker compose --env-file .env -f docker-compose.cloud.yml up -d --build core-service
docker compose -f docker-compose.cloud.yml logs --tail=200 core-service
curl http://127.0.0.1:8080/actuator/health
```

- CN: 如果你改的是 `gateway-service` 或 `profile-service`，把服务名替换掉即可。
- EN: If you changed `gateway-service` or `profile-service`, replace the service name accordingly.
- JP: `gateway-service` または `profile-service` を更新する場合は、サービス名だけ差し替えてください。

### 12.4 后端更新后必须做的验证 / Required Checks After Backend Update / 更新後に必須の確認

1. CN: 看 `docker compose logs --tail=200 <service>`，确认没有启动报错。
   EN: Review `docker compose logs --tail=200 <service>` and confirm there are no startup errors.
   JP: `docker compose logs --tail=200 <service>` を見て、起動エラーがないことを確認してください。
2. CN: 打 `http://127.0.0.1:8080/actuator/health`。
   EN: Hit `http://127.0.0.1:8080/actuator/health`.
   JP: `http://127.0.0.1:8080/actuator/health` を叩きます。
3. CN: 至少再打一个真实业务接口，例如 `mock-login` 或日记创建。
   EN: Hit at least one real business endpoint such as `mock-login` or journal creation.
   JP: `mock-login` や日記作成のような実業務 API を最低一つ叩いてください。
4. CN: 如果更新涉及数据库结构，再用 DBeaver 或 `psql` 验证表结构和数据是否符合预期。
   EN: If the update changes database structure, verify the schema and data via DBeaver or `psql`.
   JP: DB 構造変更を含む場合は、DBeaver または `psql` でスキーマとデータを確認してください。

## 13. 更新前端 / Update Frontend / フロントエンド更新

### 13.1 小程序前端更新流程 / Mini Program Frontend Update Flow / ミニプログラム更新手順

1. CN: 在本地修改 `client` 目录下的前端代码。
   EN: Modify the frontend code under `client` locally.
   JP: ローカルの `client` 配下でフロントコードを変更します。
2. CN: 本地重新打包微信小程序：
   EN: Rebuild the WeChat Mini Program locally:
   JP: ローカルで WeChat Mini Program を再ビルドします:

```bash
cd client
npm run build:weapp
```

3. CN: 在微信开发者工具重新导入或重新编译 `client/dist`。
   EN: Re-import or rebuild `client/dist` in WeChat DevTools.
   JP: WeChat DevTools で `client/dist` を再インポートまたは再ビルドします。
4. CN: 真机或开发者工具里至少再走一遍“进入记事本 -> 保存记录 -> 获取回响 -> 刷新最近记录”。
   EN: On real device or DevTools, run the path again: `enter notebook -> save record -> get resonance -> refresh recent records`.
   JP: 実機または DevTools で `記事本に入る -> 記録保存 -> 回響取得 -> 最近記録更新` をもう一度通します。
5. CN: 如果要发新体验版，在微信开发者工具点击“上传”，然后去公众平台生成新的体验版。
   EN: If you want a new trial build, click `Upload` in WeChat DevTools and generate a new trial version from the Mini Program admin console.
   JP: 新しい体験版を出すなら、WeChat DevTools で `Upload` を押し、その後ミニプログラム管理画面で新しい体験版を作成します。

### 13.2 前端更新和后端更新的区别 / Difference Between Frontend and Backend Updates / フロント更新とバック更新の違い

1. CN: 小程序前端代码不部署到腾讯云服务器，不需要去服务器上 `docker compose` 重启前端。
   EN: Mini Program frontend code is not deployed through the Tencent Cloud server, so you do not use `docker compose` on the server to restart the frontend.
   JP: ミニプログラムのフロントコードは Tencent Cloud サーバ上の `docker compose` では配布しないため、サーバでフロント再起動はしません。
2. CN: 后端更新走服务器重建容器；前端更新走本地打包、开发者工具编译、上传体验版或提审。
   EN: Backend updates go through rebuilding containers on the server; frontend updates go through local build, DevTools compile, and upload as trial or release candidate.
   JP: バックエンド更新はサーバ上でコンテナ再構築、フロント更新はローカルビルド、DevTools コンパイル、体験版アップロードまたは審査提出で進みます。

## 14. 实际踩坑记录 / Real Detours We Hit / 実際に踏んだ落とし穴

### 14.1 不要把 `docker compose ps` 当成功 / Do Not Treat `docker compose ps` as Success / `docker compose ps` を成功扱いしない

- CN: 容器全是 `Up` 只能说明进程没死，不代表业务链路通了。我们实际碰到过 `gateway-service` 健康检查正常，但业务接口因为网关路由写错直接 404。
- EN: Containers showing `Up` only means the processes did not crash. It does not prove the business path works. We actually hit a case where `gateway-service` was healthy but business routes still returned 404 because the gateway routes were wrong.
- JP: コンテナが `Up` でも、プロセスが死んでいないだけです。業務経路が通る証拠にはなりません。実際に `gateway-service` の health は正常でも、ルート設定ミスで業務 API が 404 になったことがあります。

- CN: 规避方式：每次部署完成后，至少同时验证 `actuator/health` 和一个真实业务接口，例如 `mock-login`。
- EN: Prevention: after every deployment, verify both `actuator/health` and at least one real business endpoint such as `mock-login`.
- JP: 回避策: デプロイ後は毎回 `actuator/health` だけでなく、`mock-login` のような実際の業務 API も必ず確認してください。

### 14.2 云服务器上没同步最新文件时，不要死等 / Do Not Stall When the Server Repo Is Behind / サーバ側リポジトリが古い時に止まらない

- CN: 我们实际碰到过本地仓库已经有 `server/deploy/nginx/sunyufei5.art.conf`，但云服务器上的目录没有这个文件，结果 `cp` 直接失败，后面的软链接和 reload 连锁报错。
- EN: We actually hit a case where the local repo already had `server/deploy/nginx/sunyufei5.art.conf`, but the server-side copy did not, so `cp` failed immediately and the later symlink and reload steps failed in sequence.
- JP: ローカル側には `server/deploy/nginx/sunyufei5.art.conf` があるのに、サーバ側にはまだ無く、`cp` が即失敗し、その後のシンボリックリンクと reload まで連鎖的に壊れたことがありました。

- CN: 规避方式：现场直接用 heredoc 生成配置文件，别把整个部署流程卡死在一个缺文件问题上。
- EN: Prevention: generate the config in place with a heredoc instead of blocking the whole deployment on a missing file.
- JP: 回避策: ファイル欠落で全体を止めず、heredoc でその場で設定ファイルを生成してください。

### 14.3 `.env` 一行写错，端口就全废 / One Bad `.env` Line Breaks Port Mapping / `.env` の 1 行ミスでポート設定が壊れる

- CN: 我们实际把 `GATEWAY_PORT=8080` 写成过 `808080`，Docker 直接报 `invalid hostPort: 808080`。
- EN: We actually mistyped `GATEWAY_PORT=8080` as `808080`, and Docker immediately failed with `invalid hostPort: 808080`.
- JP: 実際に `GATEWAY_PORT=8080` を `808080` と書き間違え、Docker が `invalid hostPort: 808080` を返しました。

- CN: 规避方式：改完 `.env` 后先 `cat .env` 看一眼，再执行 `docker compose up`。这种低级错误不值得靠重试去发现。
- EN: Prevention: after editing `.env`, run `cat .env` once before `docker compose up`. This kind of low-level mistake should be caught by inspection, not by retrying blindly.
- JP: 回避策: `.env` を編集したら `docker compose up` の前に一度 `cat .env` で確認してください。この種の初歩的ミスは再試行ではなく目視で潰すべきです。

### 14.4 Nginx 重载前必须先 `nginx -t` / Always Run `nginx -t` Before Reload / reload 前に必ず `nginx -t`

- CN: 我们实际遇到过 `sites-enabled` 里挂了一个指向不存在文件的软链接，结果 `nginx -t` 失败，`systemctl reload nginx` 也跟着失败。
- EN: We actually hit a broken symlink in `sites-enabled` that pointed to a missing file, so `nginx -t` failed and `systemctl reload nginx` failed right after it.
- JP: `sites-enabled` に存在しないファイルへのシンボリックリンクが残り、`nginx -t` が失敗し、その直後の `systemctl reload nginx` も失敗したことがありました。

- CN: 规避方式：顺序固定成 `nginx -t -> reload`。`nginx -t` 没过，就先删坏链接或修配置，不要硬 reload。
- EN: Prevention: keep the order fixed as `nginx -t -> reload`. If `nginx -t` fails, remove the bad link or fix the config first. Do not force a reload.
- JP: 回避策: 手順は必ず `nginx -t -> reload` に固定してください。`nginx -t` が失敗したら、壊れたリンクを削除するか設定を直してから進みます。reload を強行しないでください。

### 14.5 不要把终端提示符和输出再贴回 shell / Do Not Paste Prompts and Output Back Into the Shell / プロンプトや出力をそのまま shell に貼り戻さない

- CN: 这次部署里最浪费时间的一类错误，不是服务错，而是把 `ubuntu@...$`、日志输出、甚至半截命令一起贴回终端，结果 shell 把它们当新命令执行。
- EN: One of the biggest time-wasters in this deployment was not a broken service but pasting `ubuntu@...$`, command output, or partial logs back into the shell, causing the shell to execute garbage as commands.
- JP: 今回のデプロイで一番時間を食った原因の一つはサービス不具合ではなく、`ubuntu@...$` やログ出力、途中の文字列まで shell に貼り戻してしまい、shell がそれを新しいコマンドとして実行したことです。

- CN: 规避方式：只复制纯命令块，不复制提示符，不复制输出，不把解释文字混进 shell。
- EN: Prevention: copy only the raw command block. Do not copy the prompt, output, or explanatory text into the shell.
- JP: 回避策: shell には純粋なコマンドだけを貼り付けてください。プロンプト、出力、説明文は入れないでください。

### 14.6 正式接域名后，不要继续把网关直接裸奔到公网 / Do Not Keep the Gateway Directly Exposed After HTTPS Is Ready / HTTPS 後も Gateway を公網に裸で出し続けない

- CN: 这次云端链路先用过公网 `8080` 直连做第一阶段验证，但一旦 Nginx + HTTPS 打通，就没必要继续把 Spring 网关直接裸露在公网边缘。
- EN: In this deployment we temporarily used direct public `8080` access for the first-phase validation, but once Nginx + HTTPS is working there is no reason to keep the Spring gateway directly exposed on the public edge.
- JP: 今回は第一段階の確認として一時的に公網 `8080` 直結を使いましたが、Nginx + HTTPS が通った後は Spring Gateway をそのまま公網に晒し続ける理由はありません。

- CN: 规避方式：后续收口为 `80/443 -> Nginx -> 127.0.0.1:8080`，并在安全组里关闭公网 `8080`。
- EN: Prevention: converge to `80/443 -> Nginx -> 127.0.0.1:8080`, and close public `8080` in the security group.
- JP: 回避策: 最終的には `80/443 -> Nginx -> 127.0.0.1:8080` に収束させ、セキュリティグループでも公網 `8080` を閉じてください。
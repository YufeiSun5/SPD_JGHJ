# 测试与验证说明 / Testing and Verification Guide / テストと検証ガイド

## 1. 目标 / Goal / 目的

- CN: 这份文档定义这个项目里“测试做完了”到底是什么意思，避免把空壳 `contextLoads()` 当成功能验证。
- EN: This document defines what “testing is done” actually means in this project, so an empty `contextLoads()` test is not mistaken for real feature validation.
- JP: この文書は、このプロジェクトで「テスト完了」が何を意味するかを定義します。空の `contextLoads()` を機能検証と誤認しないためです。

- CN: 规则很直接。核心业务模块必须至少满足两件事：有最小自动化测试，且有人能按文档把功能跑通。
- EN: The rule is blunt. Every core business module must satisfy two conditions at minimum: it has minimal automated tests, and a human can execute the feature end to end by following a document.
- JP: ルールは単純明快です。主要な業務モジュールは最低でも二つを満たす必要があります。最小限の自動テストがあること、そして文書に従って人間が機能を最後まで動かせることです。

## 2. 当前适用范围 / Current Scope / 現在の適用範囲

- CN: 当前测试基线覆盖后端三服务中的实际业务部分，重点是 `profile-service` 和 `core-service`。`gateway-service` 目前主要承担路由转发，所以现阶段保留启动测试即可。
- EN: The current testing baseline covers the real business-bearing parts of the three backend services, with `profile-service` and `core-service` as the priority. `gateway-service` is mainly routing for now, so a startup test is enough at this stage.
- JP: 現在のテスト基準は、三つのバックエンドサービスのうち実際に業務を担う部分を対象にします。重点は `profile-service` と `core-service` です。`gateway-service` は現時点では主にルーティング担当なので、この段階では起動確認テストで十分です。

- CN: 前端 Taro 小程序联调不在这份文档的自动化范围内，但它必须依赖这里定义的后端验证基线。
- EN: Taro mini-program integration is not part of the automated scope of this document yet, but it must rely on the backend verification baseline defined here.
- JP: Taro ミニプログラムの結合確認は、まだこの文書の自動化対象には含めません。ただし、ここで定義するバックエンド検証基準に依存します。

## 3. 测试分层 / Test Layers / テストの層

### 3.1 单元与服务测试 / Unit and Service Tests / 単体・サービステスト

- CN: 这是第一层。直接验证业务逻辑，不靠手点接口，不靠脑补结果。
- EN: This is the first layer. It verifies business logic directly, without manual clicking or hand-waving.
- JP: これが第一層です。手動操作や脳内補完に頼らず、業務ロジックを直接検証します。

- CN: 当前已落地的最小自动化测试如下：
- EN: The currently implemented minimal automated tests are:
- JP: 現時点で実装済みの最小自動テストは以下の通りです。

1. CN: `ProfileUserServiceTest` 验证 mock 登录复用用户、首次创建用户、缺失用户返回 404。
   EN: `ProfileUserServiceTest` verifies mock-login user reuse, first-time user creation, and 404 on missing user lookup.
   JP: `ProfileUserServiceTest` は mock ログイン時の既存ユーザー再利用、初回ユーザー作成、未存在ユーザー取得時の 404 を検証します。
2. CN: `JournalApplicationServiceTest` 验证创建日记、空间共振命中、最近日记回退、无数据时空结果。
   EN: `JournalApplicationServiceTest` verifies journal creation, space resonance hit, fallback to latest journal, and empty recommendation when no data exists.
   JP: `JournalApplicationServiceTest` は日記作成、空間共鳴ヒット、最新日記へのフォールバック、データなし時の空結果を検証します。

### 3.2 启动与装配测试 / Startup and Wiring Tests / 起動・配線テスト

- CN: `contextLoads()` 这类测试不是没用，但它只能证明容器能起来，不能证明业务可用。它是保底，不是交付结论。
- EN: `contextLoads()` style tests are not useless, but they only prove the container can start. They do not prove the feature works. They are a floor, not the delivery verdict.
- JP: `contextLoads()` 系のテストは無意味ではありませんが、コンテナが起動することしか証明しません。機能が使えることの証明にはなりません。最低ラインであって、完了判定ではありません。

### 3.3 人工烟雾验证 / Manual Smoke Verification / 手動スモーク検証

- CN: 自动化测试之外，还必须保留一条人能直接执行的验证路径。当前后端链路文档见 [docs/BACKEND_SMOKE_TEST.md](docs/BACKEND_SMOKE_TEST.md)。
- EN: Besides automated tests, there must be a verification path that a human can run directly. The current backend chain document is [docs/BACKEND_SMOKE_TEST.md](docs/BACKEND_SMOKE_TEST.md).
- JP: 自動テストに加えて、人間が直接実行できる検証経路を残す必要があります。現在のバックエンド連携文書は [docs/BACKEND_SMOKE_TEST.md](docs/BACKEND_SMOKE_TEST.md) です。

## 4. 当前通过标准 / Current Pass Criteria / 現在の合格基準

- CN: 一个后端功能模块要算“可以继续往下走”，至少同时满足下面四条。
- EN: A backend feature module is only “good enough to move forward” when all four conditions below are satisfied.
- JP: バックエンド機能モジュールを「次に進める状態」と見なすには、以下の四条件をすべて満たす必要があります。

1. CN: 代码可编译。
   EN: The code compiles.
   JP: コードがコンパイルできること。
2. CN: 关键业务逻辑有最小自动化测试。
   EN: Key business logic has minimal automated tests.
   JP: 主要な業務ロジックに最小限の自動テストがあること。
3. CN: 至少有一条人工 smoke 文档可执行。
   EN: At least one manual smoke document is executable.
   JP: 少なくとも一つの手動スモーク文書が実行可能であること。
4. CN: 真跑过，而不是只写了文档和测试文件名。
   EN: It has actually been executed, not just documented or named in test files.
   JP: 文書やテスト名だけではなく、実際に走らせて確認していること。

## 5. 当前命令 / Current Commands / 現在の実行コマンド

- CN: 这个仓库当前优先使用 `server/.vfox` 下的本地 Java 和 Maven，别假设全局环境一定可用。
- EN: This repository currently prefers the local Java and Maven under `server/.vfox`; do not assume the global environment is usable.
- JP: このリポジトリでは現在 `server/.vfox` 配下のローカル Java と Maven を優先します。グローバル環境が使える前提で考えないでください。

### 5.1 运行 profile-service 测试 / Run profile-service tests / profile-service テスト実行

```powershell
$env:JAVA_HOME = 'd:/DEV_D/Spring4/20k-days-resonance/server/.vfox/sdks/java'
$env:Path = "d:/DEV_D/Spring4/20k-days-resonance/server/.vfox/sdks/java/bin;" + $env:Path
& 'd:/DEV_D/Spring4/20k-days-resonance/server/.vfox/sdks/maven/bin/mvn.cmd' -f 'd:/DEV_D/Spring4/20k-days-resonance/server/profile-service/pom.xml' test
```

### 5.2 运行 core-service 测试 / Run core-service tests / core-service テスト実行

```powershell
$env:JAVA_HOME = 'd:/DEV_D/Spring4/20k-days-resonance/server/.vfox/sdks/java'
$env:Path = "d:/DEV_D/Spring4/20k-days-resonance/server/.vfox/sdks/java/bin;" + $env:Path
& 'd:/DEV_D/Spring4/20k-days-resonance/server/.vfox/sdks/maven/bin/mvn.cmd' -f 'd:/DEV_D/Spring4/20k-days-resonance/server/core-service/pom.xml' test
```

### 5.3 连续运行两个核心模块测试 / Run both core modules / 二つの主要モジュールを連続実行

```powershell
$env:JAVA_HOME = 'd:/DEV_D/Spring4/20k-days-resonance/server/.vfox/sdks/java'
$env:Path = "d:/DEV_D/Spring4/20k-days-resonance/server/.vfox/sdks/java/bin;" + $env:Path
& 'd:/DEV_D/Spring4/20k-days-resonance/server/.vfox/sdks/maven/bin/mvn.cmd' -f 'd:/DEV_D/Spring4/20k-days-resonance/server/profile-service/pom.xml' test
& 'd:/DEV_D/Spring4/20k-days-resonance/server/.vfox/sdks/maven/bin/mvn.cmd' -f 'd:/DEV_D/Spring4/20k-days-resonance/server/core-service/pom.xml' test
```

## 6. 当前已验证结果 / Current Verified State / 現在確認済みの状態

- CN: 截至当前，会真正承载第一阶段业务逻辑的 `profile-service` 与 `core-service` 测试已经跑通。
- EN: As of now, the tests for `profile-service` and `core-service`, which carry the first-stage business logic, have been executed successfully.
- JP: 現時点で、第一段階の業務ロジックを担う `profile-service` と `core-service` のテストは実行済みで成功しています。

- CN: 这不代表整个项目已经完全验证。它只代表后端最小闭环不再是嘴上说通，而是代码和命令都跑过。
- EN: This does not mean the entire project is fully verified. It means the minimal backend loop is no longer just a claim; the code and commands have actually run.
- JP: これはプロジェクト全体が完全に検証済みという意味ではありません。最小バックエンド閉ループが、口先だけではなく、実際にコードとコマンドで確認されたという意味です。

## 7. 下一步怎么测 / What to Test Next / 次に何を検証するか

1. CN: 通过 gateway 再跑一遍 [docs/BACKEND_SMOKE_TEST.md](docs/BACKEND_SMOKE_TEST.md) 里的接口链路。
   EN: Run the API chain in [docs/BACKEND_SMOKE_TEST.md](docs/BACKEND_SMOKE_TEST.md) through the gateway.
   JP: [docs/BACKEND_SMOKE_TEST.md](docs/BACKEND_SMOKE_TEST.md) にある API チェーンを gateway 経由で実行します。
2. CN: 安装 client 依赖，验证 Taro 页面能调用后端而不是只停留在静态代码。
   EN: Install client dependencies and verify that the Taro page can call the backend instead of remaining static code.
   JP: client 依存関係を導入し、Taro ページが静的コードのままではなく、実際にバックエンドを呼べることを確認します。
3. CN: 等 BSSID 真机能力接入后，再补真实空间上下文测试，而不是永远用 mock 空间 ID 自欺欺人。
   EN: Once real BSSID capability is wired in on device, add real spatial-context tests instead of pretending the mock space ID is enough forever.
   JP: 実機で BSSID 能力が接続されたら、mock 空間 ID でごまかし続けず、本物の空間コンテキストテストを追加します。
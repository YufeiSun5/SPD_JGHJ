# [CN/EN/JP] SPD_JGHJ SCADA 业务逻辑大脑

## 1. 项目定位 (Project Positioning)
- **CN**: 本项目是一个基于 Go、Wails、Vue 3、Naive UI、MQTT、MySQL 的桌面 SCADA/IIoT 监控系统，重点是焊机与产线现场的数据采集、实时监控、报警、任务联动、工单与人员班组管理。
- **EN**: This project is a desktop SCADA/IIoT monitoring system built with Go, Wails, Vue 3, Naive UI, MQTT, and MySQL, focused on shop-floor data acquisition, real-time monitoring, alarms, task automation, work-order flow, and staff/team management.
- **JP**: 本プロジェクトは、Go、Wails、Vue 3、Naive UI、MQTT、MySQL を使ったデスクトップ SCADA/IIoT 監視システムであり、現場データ収集、リアルタイム監視、アラーム、タスク連動、工単管理、班組・人員管理を主眼とします。

## 2. 核心架构事实 (Core Architecture Facts)

### A. 桌面端调用模型 (Desktop Invocation Model)
- **CN**: 这不是以前后端 HTTP API 为主的 Web 项目。桌面前端通过 Wails 绑定直接调用 Go 的 `App` 方法，主要入口是 `window.go.main.App.*` 与生成的 `wailsjs` 包装层。
- **EN**: This is not an HTTP-API-first web application. The desktop frontend calls Go `App` methods directly through Wails binding, mainly via `window.go.main.App.*` and generated `wailsjs` wrappers.
- **JP**: これは HTTP API 主体の Web アプリではありません。デスクトップフロントは Wails のバインド経由で Go の `App` メソッドを直接呼び、主な入口は `window.go.main.App.*` と生成済み `wailsjs` ラッパです。

### B. 主数据入口是 MQTT (Primary Data Ingress Is MQTT)
- **CN**: 生产现场实时数据的主入口是 MQTT 网关，不要在改动中臆造新的 REST 通信主链路。数据库用于配置、历史数据、报警、任务、工单、班组等持久化。
- **EN**: MQTT gateways are the primary ingress for live shop-floor data. Do not invent a new REST-based primary communication path during changes. The database persists configuration, history, alarms, tasks, work orders, teams, and related data.
- **JP**: 現場リアルタイムデータの主入口は MQTT ゲートウェイです。変更時に新しい REST 主経路を勝手に導入してはいけません。DB は設定、履歴、アラーム、タスク、工単、班組などの永続化に使います。

### C. Wails 启动时会连带初始化后端内核 (Wails Boot Also Initializes Backend Core)
- **CN**: `desktop/main.go` 在启动桌面窗口前会初始化数据库、TagManager、各类 Worker、任务调度器、网关管理器和热重载，不是“前端单独跑 + 后端另开 API 服务”的结构。
- **EN**: `desktop/main.go` initializes the database, TagManager, worker pools, task scheduler, gateway manager, and hot reload before opening the desktop window. This is not a “frontend only + separate API service” structure.
- **JP**: `desktop/main.go` はデスクトップウィンドウ起動前に DB、TagManager、各種 Worker、タスクスケジューラ、ゲートウェイ管理、ホットリロードを初期化します。つまり「フロント単体起動 + 別 API サービス」構成ではありません。

### D. AI 能力是旁路服务，不是主控制链路 (AI Is Auxiliary, Not the Main Control Path)
- **CN**: `desktop/ai_client.go` 通过本机 `http://127.0.0.1:8006` 调用 FastAPI 知识库服务。它属于辅助问答能力，不是主 SCADA 数据链路。
- **EN**: `desktop/ai_client.go` calls a local FastAPI knowledge service at `http://127.0.0.1:8006`. It is an auxiliary capability, not the main SCADA data path.
- **JP**: `desktop/ai_client.go` はローカル `http://127.0.0.1:8006` の FastAPI 知識サービスを呼びます。補助的な問答機能であり、主 SCADA データ経路ではありません。

## 3. 并发与数据流约束 (Concurrency And Dataflow Rules)
- **CN**: 本系统遵循 memory-first 和 channel 解耦设计。MQTT 收包、逻辑计算、报警、历史存储、任务触发、事件执行是分层异步处理，改动时不要把慢 IO 塞回 MQTT 回调或核心逻辑热路径。
- **EN**: The system follows a memory-first, channel-decoupled design. MQTT ingress, logic evaluation, alarms, history storage, task triggering, and event execution are layered asynchronous stages. Do not push slow I/O back into MQTT callbacks or the hot path.
- **JP**: 本システムは memory-first と Channel 分離設計を採用しています。MQTT 受信、ロジック評価、アラーム、履歴保存、タスク起動、イベント実行は非同期レイヤで分離されています。低速 I/O を MQTT コールバックや中核ホットパスへ戻してはいけません。

- **CN**: `docs/项目架构总览.md` 中描述的双循环快照、Worker 池、TaskScheduler、EventProcessor、GatewayManager 是当前实现思路。除非明确重构，不要破坏这些边界。
- **EN**: The dual-pass snapshot model, worker pools, TaskScheduler, EventProcessor, and GatewayManager described in `docs/项目架构总览.md` reflect the current implementation intent. Do not break these boundaries unless a refactor is explicitly required.
- **JP**: `docs/项目架构总览.md` に記載された二段スナップショット、Worker プール、TaskScheduler、EventProcessor、GatewayManager は現行設計意図です。明示的なリファクタでない限り、この境界を壊してはいけません。

## 4. 工程写作与修改规则 (Engineering Writing And Change Rules)
- **CN**: 只要是实际承载业务功能的位置，例如 Wails 绑定方法、核心 Worker、网关管理、任务调度、核心页面、关键 SQL，都必须补充中英日三语注释或说明。
- **EN**: Any location carrying real business functionality, such as Wails-bound methods, core workers, gateway management, task scheduling, core pages, and critical SQL, must include trilingual comments or descriptions in Chinese, English, and Japanese.
- **JP**: Wails バインドメソッド、コア Worker、ゲートウェイ管理、タスクスケジューリング、主要ページ、重要 SQL など、実際の業務機能を担う箇所には中英日三言語の注釈または説明を付けること。

- **CN**: 代码和注释风格默认直接、具体、少废话。先把调用链、数据源、状态流转写清楚，再考虑形式上的漂亮。
- **EN**: Code and comments should stay direct, concrete, and low-fluff. Make the call chain, data source, and state transitions explicit first; polish comes second.
- **JP**: コードとコメントは直接的・具体的・簡潔であること。呼び出し経路、データ源、状態遷移を先に明確化し、体裁はその後です。

- **CN**: 功能改动后必须保留最小可执行验证路径。桌面功能至少说明如何在 `desktop` 目录用 `wails dev` 或相关页面人工验证；数据库和 SQL 改动至少给出可复查路径。
- **EN**: After feature changes, keep a minimal executable validation path. For desktop features, at minimum describe how to validate via `wails dev` under `desktop` or by visiting the relevant page. For DB and SQL changes, provide a reviewable verification path.
- **JP**: 機能変更後は最小限の実行可能な検証経路を残すこと。デスクトップ機能なら `desktop` 配下での `wails dev` や対象画面での手動確認方法を示し、DB/SQL 変更には再確認可能な検証経路を残すこと。

## 5. 当前项目状态 (Current Project State)
- **CN**: 当前桌面主栈为 Go + Wails v2 + Vue 3 + Vite + Naive UI。前端路由使用 Hash 模式，以适配桌面壳路径环境。
- **EN**: The current desktop stack is Go + Wails v2 + Vue 3 + Vite + Naive UI. The frontend router uses hash history to fit the desktop shell path environment.
- **JP**: 現在のデスクトップ主スタックは Go + Wails v2 + Vue 3 + Vite + Naive UI です。フロントルータはデスクトップシェル向けに Hash モードを使います。

- **CN**: 当前前端核心页面已覆盖驾驶舱、生产、质量、报警、人员、设备状态、历史、设备配置、任务管理、系统设置、AI 助手等业务模块。
- **EN**: The frontend currently covers cockpit, production, quality, alarms, staff, device status, history, device configuration, task management, system settings, and AI assistant modules.
- **JP**: 現在のフロントはコックピット、生産、品質、アラーム、人員、設備状態、履歴、設備設定、タスク管理、システム設定、AI アシスタントなどの主要モジュールを備えています。

- **CN**: 当前 `config.env` 默认数据库为 `localhost:3306 / spd_jghj`。代码中仍保留通用数据库字段命名和历史 PostgreSQL 文案，但实际依赖包含 MySQL 驱动，MQTT 多网关配置的真实来源已转向数据库表，不再以 `config.env` 中的 MQTT 配置为主。
- **EN**: `config.env` currently defaults to `localhost:3306 / spd_jghj` for the database. The code still contains generic DB naming and some historical PostgreSQL wording, but the runtime depends on the MySQL driver. Multi-gateway MQTT configuration now comes primarily from database tables rather than `config.env`.
- **JP**: `config.env` の現行デフォルト DB は `localhost:3306 / spd_jghj` です。コードには汎用的な DB 命名や過去の PostgreSQL 文言が残っていますが、実行時依存は MySQL ドライバです。MQTT マルチゲートウェイ設定の正本は `config.env` ではなく DB テーブル側へ移っています。

## 6. 哪些内容写进长期规则，哪些写成流程文档 (Rules Vs Workflow Docs)
- **CN**: 长期稳定、默认必须遵守的内容写进 `AGENTS.md`，例如 Wails 直绑调用模型、MQTT 主入口、并发边界、三语注释要求、修改后的最低验证标准。
- **EN**: Put long-lived default rules into `AGENTS.md`, such as the Wails direct-binding model, MQTT-first ingress, concurrency boundaries, trilingual comment requirements, and minimum post-change validation standards.
- **JP**: Wails 直接バインドモデル、MQTT 主入口、並行処理の境界、三言語注釈要件、変更後の最低検証基準のような長期ルールは `AGENTS.md` に置くこと。

- **CN**: 可重复执行、步骤明确的内容写进操作文档，例如数据库初始化、网关配置、热重载检查、Wails 构建、桌面打包、SQL 修复步骤。
- **EN**: Put repeatable step-based workflows into operational docs, such as DB initialization, gateway configuration, hot-reload checks, Wails build, desktop packaging, and SQL repair procedures.
- **JP**: DB 初期化、ゲートウェイ設定、ホットリロード確認、Wails ビルド、デスクトップ配布物作成、SQL 修復手順のような、反復可能で手順が明確な内容は運用ドキュメントへ置くこと。

## 7. 通用执行规则 (Tool-Agnostic Execution Rules)

### 动手前必读（强制，不可跳过）
- **CN**: 每次开始任何代码改动、功能新增、文档更新之前，必须先完成以下两步：
  1. 读取 `MEMORY.md` 全文，了解当前项目进度、已落地事项、已知约束；
  2. 读取 `instructions/` 目录下与本次改动相关的规范文件（后端改动读 `backend-go-wails-scada.instructions.md`，前端改动读 `frontend-wails-vue.instructions.md`，代码风格与注释读 `code-style.instructions.md`）。
  未完成以上两步不得开始修改代码或文件。
- **EN**: Before starting any code change, feature addition, or documentation update, you MUST first:
  1. Read `MEMORY.md` in full to understand current project progress, completed items, and known constraints;
  2. Read the relevant files under `instructions/` for this change (backend → `backend-go-wails-scada.instructions.md`; frontend → `frontend-wails-vue.instructions.md`; style/comments → `code-style.instructions.md`).
  Do NOT begin modifying code or files until both steps are done.
- **JP**: コード変更・機能追加・ドキュメント更新を始める前に、必ず次の2ステップを完了すること：
  1. `MEMORY.md` を全文読み、現在のプロジェクト進捗・完了事項・既知の制約を把握する；
  2. `instructions/` 配下の関連ファイルを読む（バックエンド変更 → `backend-go-wails-scada.instructions.md`、フロント変更 → `frontend-wails-vue.instructions.md`、スタイル/注釈 → `code-style.instructions.md`）。
  この2ステップを完了するまでコードやファイルの変更を開始してはならない。

- **CN**: 本仓库的长期事实来源以 `AGENTS.md`、`MEMORY.md`、`docs/项目架构总览.md`、`docs/启动说明.md` 及核心代码实现为准。旧 README 中若存在与当前桌面架构不一致的内容，应优先相信当前实现而不是旧说明。
- **EN**: Long-lived repository truth comes from `AGENTS.md`, `MEMORY.md`, `docs/项目架构总览.md`, `docs/启动说明.md`, and the core implementation. If the old README disagrees with the current desktop architecture, trust the current implementation over stale prose.
- **JP**: 本リポジトリの長期的な正本は `AGENTS.md`、`MEMORY.md`、`docs/项目架构总览.md`、`docs/启动说明.md`、およびコア実装です。古い README が現行デスクトップ構成と食い違う場合、古い説明ではなく現行実装を優先すること。

- **CN**: 修改前端时，优先复用已有 `window.go.main.App` / `wailsjs` 能力，不要先入为主补 HTTP API。只有在明确需要拆分桌面与服务端边界时，才引入新的接口层。
- **EN**: When changing the frontend, prefer existing `window.go.main.App` / `wailsjs` capabilities. Do not add HTTP APIs by assumption. Introduce a new interface layer only when a real desktop/service split is explicitly required.
- **JP**: フロント変更時は既存の `window.go.main.App` / `wailsjs` を優先利用し、思い込みで HTTP API を追加しないこと。デスクトップとサービスの境界分離が明確に必要な場合にのみ、新しいインターフェース層を導入すること。

- **CN**: 阶段性工作完成后，必须主动回看是否需要同步更新 `AGENTS.md`、`MEMORY.md` 和相关说明文档。新增了长期约束、稳定流程、项目现状变化时，不要只改代码不更新知识资产。
- **EN**: After a meaningful stage of work, proactively check whether `AGENTS.md`, `MEMORY.md`, and related docs need updates. If the change introduces long-lived rules, stable workflows, or project-state shifts, update the knowledge assets along with the code.
- **JP**: ひと区切りの作業後は、`AGENTS.md`、`MEMORY.md`、関連ドキュメントの更新要否を必ず見直すこと。長期ルール、安定運用手順、現状の変化が入ったなら、コードだけで終わらせず知識資産も更新すること。
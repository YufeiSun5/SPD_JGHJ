# SPD_JGHJ 文档索引

本目录用于收纳当前 Wails + Vue + Go SCADA 项目的业务说明、操作说明、架构解释和人工验证文档。

## 当前文档分组

1. 架构与索引
   - PROJECT_CONTEXT.md
   - 项目架构总览.md
   - 系统文件功能说明.md
   - Go并发架构优势.md
2. 启动与验证
   - BACKEND_SMOKE_TEST.md
   - 启动说明.md
   - 安装依赖.md
   - 命令集合.md
3. 业务功能说明
   - 驾驶舱使用说明.md
   - 任务管理功能说明.md
   - 任务去重机制说明.md
   - AI知识问答系统架构说明.md
4. 配置与数据说明
   - mqtt_topic_mapping.md
   - 质量码监控配置说明.md
   - 设备空闲状态检测配置.md
   - test_mqtt.md
   - 使用示例_FALSE_TO_TRUE任务.md

## 使用原则

1. 长期稳定规则写入 AGENTS.md 或 instructions/。
2. 可重复执行的步骤写入 skills/。
3. 根目录不再堆放普通业务 Markdown，默认收口到 docs/。
4. 发现文档与实际代码冲突时，优先相信当前代码实现，再回头修正文档。

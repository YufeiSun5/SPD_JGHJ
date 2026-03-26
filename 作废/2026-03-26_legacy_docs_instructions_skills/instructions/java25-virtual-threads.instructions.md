# [Java 虚拟线程开发指令]
# - 当前仓库实际运行基线是 Java 21，不要默认假设本地已经在 Java 25。
# - Java 21 已正式支持虚拟线程；如需启用 Spring 虚拟线程，再显式增加 spring.threads.virtual.enabled=true。
# - 当前后端先以稳定启动和调试为优先，未默认打开虚拟线程配置。
# - 规范: 避免滥用 ThreadLocal；如后续升级更高版本 Java，可再评估 Scoped Values / StructuredTaskScope。
# - 修改并发模型前，先确认 JDBC、Redis、外部 AI 调用链路的阻塞行为与线程数量预期。

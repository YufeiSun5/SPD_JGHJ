# [技能] 20k-days-resonance 后端基础约束

- 当前后端为轻量三服务结构: gateway-service、core-service、profile-service。
- 本地开发环境以 Windows + Docker + vfox 为准，不依赖云主机做日常开发调试。
- Java 使用 server/.vfox/sdks/java 的 Java 21；Maven 使用 server/.vfox/sdks/maven/bin/mvn.cmd。
- PostgreSQL 拆分为 resonance_profile 与 resonance_core；gateway-service 不持库。
- Redis 为本地 Docker 单实例，主要用于空间共振热点 weekday 缓存。
- AI 分析当前属于 core 业务数据的一部分，不单独拆数据库。
- 修改表结构时，优先维护 sql/profile 或 sql/core 下的 create/modify/test 版本化脚本。
- 不要在 core 与 profile 之间建立跨数据库外键。
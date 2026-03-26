# [Instruction] 空间共振功能实现指南

## 前端任务 (Taro)
1. 获取 BSSID: 调用 `Taro.getConnectedWifi`。
2. 封装 Service: 发送请求时 Header 携带 `X-Space-Id` (SHA256(BSSID))。
3. 状态管理: 使用 Context 存储当前空间活跃的“礼拜几”。

## 后端任务 (Spring Boot)
1. Redis 设计: `opsForValue().set("res:" + hashedBssid, weekday, 1, HOURS)`。
2. 数据库查询: `findByUserIdAndOriginalWeekday`。
3. 异常处理: 若 BSSID 获取不到，降级为普通的随机推荐（AI 语义推荐）。
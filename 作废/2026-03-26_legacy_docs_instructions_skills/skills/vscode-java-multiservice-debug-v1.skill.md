# [技能] VS Code 多 Java 服务调试避坑

- 多项目工作区下，Java attach 调试必须补 projectName，否则源码定位、表达式求值、断点命中表现会不稳定。
- attach 配置建议同时补 sourcePaths，避免调用堆栈停住但编辑器不自动跳到源码。
- 想命中启动期断点时，JDWP 参数必须使用 suspend=y；suspend=n 会让 main 方法在 attach 前直接跑过去。
- Windows + VS Code tasks + Maven 长跑进程组合容易导致 attach 超时，因此当前采用先后台启动，再 attach 的方式。
- 调试日志写入 .vscode/logs，文件名必须带秒级时间戳，避免多次启动相互覆盖。
- 若调试端口已监听但 VS Code 报 attach timeout，先检查服务是否已实际启动，再决定是重启还是直接使用 Attach 配置。
- 多服务并行调试时，优先使用 launch.json 里的独立配置或 compound，而不是在编辑器 CodeLens 上直接逐个点 Debug。
# Function Calling 架构详解

## 📋 目录
1. [决策流程：何时使用函数调用](#决策流程)
2. [文件协作关系](#文件协作关系)
3. [如何扩展新方法](#如何扩展新方法)
4. [如何废弃旧方法](#如何废弃旧方法)
5. [完整调用链路](#完整调用链路)

---

## 🎯 决策流程：何时使用函数调用

### 第一步：模式判断（config.py）
```
用户提问 → knowledge_service.query_knowledge_stream()
         ↓
    检查 config.PURE_RAG_MODE
         ↓
    ├─ True  → 直接走纯RAG流程（跳过函数调用）
    └─ False → 进入智能模式（可能使用函数调用）
```

**配置位置**：`config.py` 第 24 行
```python
PURE_RAG_MODE = False  # False=智能模式, True=纯RAG模式
```

### 第二步：AI 意图识别（智能模式下）

**决策者**：`services/ai_function_caller.py` 或 `services/function_caller.py`

#### 方案A：AI Function Calling（推荐，更智能）
**文件**：`services/ai_function_caller.py`
**决策位置**：第 118-151 行

```
用户问题 → analyze_and_call()
         ↓
    构建 Function Calling Prompt（包含所有可用函数定义）
         ↓
    LLM 分析问题 → 输出 JSON 决策
         ↓
    ├─ function_call: null → 不需要函数调用，回退到 RAG
    └─ function_call: {name, arguments} → 执行指定函数
```

**AI 决策示例**：
```json
// 需要调用函数
{
    "function_call": {
        "name": "query_defect_rate",
        "arguments": {
            "date": "2024-12-05",
            "model_name": "iPhone15"
        }
    },
    "reasoning": "用户询问今天iPhone15的不良率，需要查询数据库"
}

// 不需要调用函数
{
    "function_call": null,
    "reasoning": "用户问的是通用知识，不需要查询数据库"
}
```

#### 方案B：传统 Function Calling（基于规则）
**文件**：`services/function_caller.py`
**决策位置**：第 225-280 行

```
用户问题 → analyze_and_call()
         ↓
    关键词匹配 + 正则表达式
         ↓
    ├─ 匹配到数据库查询关键词 → 选择对应函数
    └─ 未匹配到 → 不调用函数，回退到 RAG
```

**关键词匹配规则**：
- "不良率"、"质量" → `query_defect_rate`
- "生产统计" → `query_production_stats`
- "计划"、"安排" → `query_production_plan`
- "效率"、"生产线" → `query_production_line_efficiency`

### 第三步：函数执行
**执行者**：`services/function_registry.py`
**位置**：第 471-517 行

```
execute_function(function_name, parameters)
         ↓
    1. 参数解析和验证（parse_parameters）
         ↓
    2. 执行 SQL 查询（使用 sql_template）
         ↓
    3. 返回结构化数据
```

---

## 🔗 文件协作关系

```
┌─────────────────────────────────────────────────────────────┐
│                        用户请求入口                          │
│                   routers/knowledge.py                       │
│                  POST /api/knowledge/query                   │
└──────────────────────────┬──────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│                      核心服务层                              │
│              services/knowledge_service.py                   │
│         query_knowledge_stream() - 第 623-750 行            │
│                                                              │
│  ┌─────────────────────────────────────────────────┐       │
│  │ 1. 检查 config.PURE_RAG_MODE                    │       │
│  │    ├─ True  → _pure_rag_query_stream()         │       │
│  │    └─ False → 继续智能模式                      │       │
│  │                                                  │       │
│  │ 2. 尝试 AI Function Calling                     │       │
│  │    调用 ai_function_caller.analyze_and_call()   │       │
│  │                                                  │       │
│  │ 3. 回退到传统 Function Calling                  │       │
│  │    调用 function_caller.analyze_and_call()      │       │
│  │                                                  │       │
│  │ 4. 最终回退到纯 RAG                             │       │
│  └─────────────────────────────────────────────────┘       │
└──────────────────────────┬──────────────────────────────────┘
                           ↓
         ┌─────────────────┴─────────────────┐
         ↓                                     ↓
┌─────────────────────┐           ┌─────────────────────┐
│   AI Function Caller │           │ Traditional Caller  │
│ ai_function_caller.py│           │ function_caller.py  │
│                      │           │                     │
│ • LLM 分析意图       │           │ • 关键词匹配        │
│ • 智能参数提取       │           │ • 正则表达式提取    │
│ • 自适应决策         │           │ • 规则引擎          │
└──────────┬───────────┘           └──────────┬──────────┘
           │                                   │
           └─────────────┬─────────────────────┘
                         ↓
         ┌───────────────────────────────────┐
         │      函数注册中心（核心）          │
         │   services/function_registry.py   │
         │                                    │
         │  ┌──────────────────────────────┐ │
         │  │ 1. 函数定义存储              │ │
         │  │    functions: Dict[str, DB]  │ │
         │  │                              │ │
         │  │ 2. 参数解析和验证            │ │
         │  │    parse_parameters()        │ │
         │  │                              │ │
         │  │ 3. SQL 执行                  │ │
         │  │    execute_function()        │ │
         │  └──────────────────────────────┘ │
         └───────────────┬───────────────────┘
                         ↓
         ┌───────────────────────────────────┐
         │      函数定义文件（可选）          │
         │ services/function_definitions.py  │
         │                                    │
         │ • OpenAI 格式函数定义             │
         │ • 用于标准 Function Calling       │
         │ • 提供日期上下文辅助              │
         └───────────────────────────────────┘
                         ↓
         ┌───────────────────────────────────┐
         │      函数管理接口（可选）          │
         │  routers/function_management.py   │
         │                                    │
         │ • GET  /api/functions             │
         │ • POST /api/functions/register    │
         │ • POST /api/functions/test        │
         │ • DELETE /api/functions/{name}    │
         └───────────────────────────────────┘
```

### 关键配置文件
```
config.py
├─ PURE_RAG_MODE: bool          # 模式开关
├─ DEBUG_FUNCTION_CALLING: bool # 调试开关
└─ DATABASE_URL: str            # 数据库连接
```

---

## 🚀 如何扩展新方法

### 方法1：通过 API 动态注册（推荐，无需重启）

**接口**：`POST /api/functions/register`

**示例：添加"设备故障查询"函数**

```json
{
    "name": "query_device_failures",
    "description": "查询指定时间范围内的设备故障记录",
    "sql_template": "SELECT device_id, failure_time, failure_type, repair_status FROM device_failures WHERE failure_time BETWEEN :start_date AND :end_date AND (:device_id IS NULL OR device_id = :device_id) ORDER BY failure_time DESC",
    "category": "maintenance",
    "parameters": [
        {
            "name": "start_date",
            "type": "date",
            "description": "开始日期，格式：YYYY-MM-DD",
            "required": true,
            "validation_regex": "^\\d{4}-\\d{2}-\\d{2}$"
        },
        {
            "name": "end_date",
            "type": "date",
            "description": "结束日期，格式：YYYY-MM-DD",
            "required": true,
            "validation_regex": "^\\d{4}-\\d{2}-\\d{2}$"
        },
        {
            "name": "device_id",
            "type": "string",
            "description": "设备ID（可选），如：DEV-001",
            "required": false,
            "default_value": null
        }
    ],
    "examples": [
        "查询本周的设备故障",
        "DEV-001最近有哪些故障",
        "本月设备故障统计"
    ]
}
```

**优点**：
- ✅ 无需修改代码
- ✅ 无需重启服务
- ✅ 支持热更新
- ✅ 可通过管理界面操作

### 方法2：代码级注册（需要重启）

**文件**：`services/function_registry.py`
**位置**：`_initialize_builtin_functions()` 方法内（第 84-363 行）

**步骤**：

#### 1. 定义函数对象
```python
# 在 _initialize_builtin_functions() 方法中添加
device_failure_function = DatabaseFunction(
    name="query_device_failures",
    description="查询设备故障记录",
    sql_template="""
        SELECT 
            device_id,
            failure_time,
            failure_type,
            repair_status,
            repair_time
        FROM device_failures 
        WHERE failure_time BETWEEN :start_date AND :end_date
            AND (:device_id IS NULL OR device_id = :device_id)
        ORDER BY failure_time DESC
    """,
    parameters=[
        FunctionParameter(
            name="start_date",
            type=ParameterType.DATE,
            description="开始日期，格式：YYYY-MM-DD",
            validation_regex=r"^\d{4}-\d{2}-\d{2}$"
        ),
        FunctionParameter(
            name="end_date",
            type=ParameterType.DATE,
            description="结束日期，格式：YYYY-MM-DD",
            validation_regex=r"^\d{4}-\d{2}-\d{2}$"
        ),
        FunctionParameter(
            name="device_id",
            type=ParameterType.STRING,
            description="设备ID（可选）",
            required=False,
            default_value=None
        )
    ],
    category="maintenance",
    examples=[
        "查询本周的设备故障",
        "DEV-001最近有哪些故障"
    ]
)
```

#### 2. 注册函数
```python
# 在 _initialize_builtin_functions() 方法末尾添加
self.register_function(device_failure_function)
```

#### 3. （可选）更新 AI Function Definitions
**文件**：`services/function_definitions.py`
**位置**：`get_function_definitions()` 函数内

```python
# 在返回的列表中添加
{
    "type": "function",
    "function": {
        "name": "query_device_failures",
        "description": "查询设备故障记录，包括故障时间、类型和维修状态",
        "parameters": {
            "type": "object",
            "properties": {
                "start_date": {
                    "type": "string",
                    "description": "开始日期，格式为YYYY-MM-DD",
                    "pattern": "^\\d{4}-\\d{2}-\\d{2}$"
                },
                "end_date": {
                    "type": "string",
                    "description": "结束日期，格式为YYYY-MM-DD",
                    "pattern": "^\\d{4}-\\d{2}-\\d{2}$"
                },
                "device_id": {
                    "type": "string",
                    "description": "设备ID（可选）"
                }
            },
            "required": ["start_date", "end_date"]
        }
    }
}
```

#### 4. （可选）更新传统 Function Caller 的关键词匹配
**文件**：`services/function_caller.py`
**位置**：`FunctionMatcher` 类的 `match_function()` 方法

```python
# 添加关键词匹配规则
if any(keyword in query_lower for keyword in ['设备故障', '故障记录', 'device failure']):
    return 'query_device_failures'
```

### 扩展性评估

#### ✅ 优点：
1. **高度模块化**：函数定义与执行逻辑分离
2. **灵活配置**：支持动态注册和代码注册两种方式
3. **参数验证**：自动类型转换和格式验证
4. **AI 友好**：函数定义可直接被 LLM 理解
5. **易于测试**：提供专门的测试接口

#### ⚠️ 注意事项：
1. **SQL 注入风险**：必须使用参数化查询（`:param` 格式）
2. **性能考虑**：复杂 SQL 可能影响响应速度
3. **权限控制**：目前没有函数级权限管理
4. **数据库依赖**：所有函数都依赖同一个数据库连接

---

## 🗑️ 如何废弃旧方法

### 方案1：软删除（推荐，可回滚）

#### 步骤1：标记为废弃
**文件**：`services/function_registry.py`

```python
# 在 DatabaseFunction 类中添加 deprecated 字段
@dataclass
class DatabaseFunction:
    name: str
    description: str
    sql_template: str
    parameters: List[FunctionParameter]
    category: str = "query"
    examples: List[str] = None
    deprecated: bool = False  # 新增字段
    deprecated_reason: str = ""  # 废弃原因
    replacement_function: str = ""  # 替代函数
```

#### 步骤2：修改函数定义
```python
# 标记旧函数为废弃
old_function = DatabaseFunction(
    name="query_old_defect_rate",  # 旧函数
    description="【已废弃】请使用 query_defect_rate 代替",
    sql_template="...",
    parameters=[...],
    deprecated=True,
    deprecated_reason="此函数已被 query_defect_rate 替代，提供更好的性能",
    replacement_function="query_defect_rate"
)
```

#### 步骤3：添加废弃警告
**文件**：`services/function_registry.py`
**位置**：`execute_function()` 方法

```python
async def execute_function(self, function_name: str, parameters: Dict[str, Any]) -> Dict[str, Any]:
    function = self.get_function(function_name)
    if not function:
        raise ValueError(f"未找到函数: {function_name}")
    
    # 检查是否废弃
    if function.deprecated:
        warning_msg = f"⚠️ 警告：函数 {function_name} 已废弃"
        if function.deprecated_reason:
            warning_msg += f" - {function.deprecated_reason}"
        if function.replacement_function:
            warning_msg += f" - 请使用 {function.replacement_function} 代替"
        encoding_fix.safe_print(warning_msg)
    
    # 继续执行...
```

#### 步骤4：从 AI 函数列表中移除
**文件**：`services/function_definitions.py`

```python
def get_function_definitions() -> List[Dict[str, Any]]:
    """获取所有可用函数的OpenAI格式定义"""
    
    # 从 function_registry 获取所有未废弃的函数
    from services.function_registry import function_registry
    
    active_functions = [
        func for func in function_registry.list_functions()
        if not func.deprecated  # 过滤掉废弃函数
    ]
    
    # 转换为 OpenAI 格式
    return [convert_to_openai_format(func) for func in active_functions]
```

### 方案2：硬删除（不可回滚）

#### 通过 API 删除
```bash
DELETE /api/functions/query_old_function
```

**限制**：内置函数无法通过 API 删除（见 `routers/function_management.py` 第 285-290 行）

#### 代码级删除
**文件**：`services/function_registry.py`

```python
# 在 _initialize_builtin_functions() 中
# 1. 注释掉函数定义
# old_function = DatabaseFunction(...)

# 2. 注释掉注册调用
# self.register_function(old_function)
```

**同时删除**：
1. `services/function_definitions.py` 中的 OpenAI 格式定义
2. `services/function_caller.py` 中的关键词匹配规则

### 方案3：版本化管理（企业级）

#### 目录结构
```
services/
├─ function_registry.py          # 核心注册器
├─ functions/
│  ├─ v1/
│  │  ├─ quality_functions.py    # 质量相关函数 v1
│  │  └─ production_functions.py # 生产相关函数 v1
│  ├─ v2/
│  │  ├─ quality_functions.py    # 质量相关函数 v2（改进版）
│  │  └─ production_functions.py
│  └─ deprecated/
│     └─ old_functions.py        # 废弃函数存档
```

#### 版本切换配置
**文件**：`config.py`

```python
# 函数版本配置
FUNCTION_VERSION = "v2"  # 可选：v1, v2

# 是否允许调用废弃函数（用于兼容性）
ALLOW_DEPRECATED_FUNCTIONS = False
```

### 废弃流程最佳实践

```
1. 标记废弃（第1周）
   ├─ 添加 deprecated=True
   ├─ 设置 replacement_function
   └─ 记录日志和警告

2. 通知期（第2-4周）
   ├─ 在响应中返回废弃警告
   ├─ 文档更新
   └─ 通知使用方迁移

3. 限制使用（第5-6周）
   ├─ 从 AI 函数列表移除
   ├─ 仅允许直接调用（不推荐）
   └─ 增加错误日志

4. 完全移除（第7周+）
   ├─ 从注册表删除
   ├─ 代码归档
   └─ 更新文档
```

---

## 🔄 完整调用链路示例

### 示例：用户询问"今天iPhone15的不良率是多少？"

```
1. 用户请求
   POST /api/knowledge/query
   Body: {"question": "今天iPhone15的不良率是多少？"}
   ↓

2. 路由层（routers/knowledge.py）
   knowledge.router → query_knowledge()
   ↓

3. 服务层（services/knowledge_service.py）
   knowledge_service.query_knowledge_stream()
   ├─ 检查 config.PURE_RAG_MODE = False（智能模式）
   └─ 调用 ai_function_caller.analyze_and_call()
   ↓

4. AI 意图分析（services/ai_function_caller.py）
   analyze_and_call("今天iPhone15的不良率是多少？")
   ├─ 构建 Prompt（包含所有函数定义）
   ├─ LLM 分析 → 输出：
   │   {
   │     "function_call": {
   │       "name": "query_defect_rate",
   │       "arguments": {
   │         "date": "2024-12-05",
   │         "model_name": "iPhone15"
   │       }
   │     },
   │     "reasoning": "用户询问今天iPhone15的不良率"
   │   }
   └─ 调用 _execute_function()
   ↓

5. 函数执行（services/function_registry.py）
   function_registry.execute_function("query_defect_rate", {...})
   ├─ 获取函数定义
   ├─ 解析参数：parse_parameters()
   │   ├─ date: "2024-12-05" (DATE 类型，验证格式)
   │   └─ model_name: "iPhone15" (STRING 类型)
   ├─ 执行 SQL：
   │   SELECT date, model_name, total_count, defect_count,
   │          ROUND((defect_count * 100.0 / total_count), 2) as defect_rate
   │   FROM production_quality 
   │   WHERE date = '2024-12-05' AND model_name = 'iPhone15'
   └─ 返回结果：
       {
         "success": true,
         "function_name": "query_defect_rate",
         "data": [
           {
             "date": "2024-12-05",
             "model_name": "iPhone15",
             "total_count": 1000,
             "defect_count": 25,
             "defect_rate": 2.5
           }
         ],
         "count": 1
       }
   ↓

6. 结果处理（services/knowledge_service.py）
   ├─ 格式化数据库结果
   ├─ 构建上下文 Prompt
   ├─ LLM 生成自然语言回答：
   │   "根据查询结果，今天（2024年12月5日）iPhone15的生产情况如下：
   │    - 总产量：1000台
   │    - 不良品数：25台
   │    - 不良率：2.5%
   │    
   │    不良率控制在合理范围内。"
   └─ 流式返回给用户
   ↓

7. 响应返回
   SSE Stream → 前端逐字显示
```

### 如果是纯 RAG 模式（PURE_RAG_MODE=True）

```
1-2. [同上]
   ↓

3. 服务层（services/knowledge_service.py）
   knowledge_service.query_knowledge_stream()
   ├─ 检查 config.PURE_RAG_MODE = True（纯RAG模式）
   └─ 直接调用 _pure_rag_query_stream()
   ↓

4. 纯 RAG 流程
   _pure_rag_query_stream("今天iPhone15的不良率是多少？")
   ├─ 向量检索：从 ChromaDB 查询相关文档
   ├─ 构建上下文：整合检索到的知识
   ├─ LLM 生成回答：基于知识库内容
   └─ 流式返回
   ↓

5. 响应返回
   SSE Stream → 前端逐字显示
   
   注意：此模式下不会查询数据库，只基于知识库文档回答
```

---

## 📊 两种模式对比

| 特性 | 纯 RAG 模式 | 智能模式（RAG + Function Calling） |
|------|-------------|-----------------------------------|
| **配置** | `PURE_RAG_MODE = True` | `PURE_RAG_MODE = False` |
| **数据源** | 仅知识库文档（ChromaDB） | 知识库 + 实时数据库查询 |
| **响应速度** | 快（无数据库查询） | 较慢（需数据库查询） |
| **数据时效性** | 取决于知识库更新 | 实时数据 |
| **适用场景** | 通用知识问答、文档查询 | 数据分析、报表查询、实时监控 |
| **AI 能力** | 文本理解和生成 | 意图识别 + 工具调用 + 文本生成 |
| **扩展性** | 添加文档 | 添加文档 + 添加函数 |
| **复杂度** | 低 | 高 |
| **可控性** | 高（仅依赖知识库） | 中（依赖 AI 判断） |

---

## 🎓 总结

### 核心设计理念
1. **分层架构**：路由 → 服务 → 函数注册 → 数据库
2. **模式切换**：通过配置灵活切换 RAG 和 Function Calling
3. **AI 驱动**：让 LLM 决定何时调用函数，而非硬编码规则
4. **高扩展性**：支持动态注册新函数，无需修改核心代码
5. **向后兼容**：废弃函数可以软删除，保留兼容性

### 扩展建议
1. **添加新函数**：优先使用 API 动态注册
2. **废弃旧函数**：使用软删除 + 通知期策略
3. **性能优化**：为常用函数添加缓存
4. **安全加固**：添加函数级权限控制
5. **监控告警**：记录函数调用日志和性能指标

### 可拓展性评分
- **易用性**：⭐⭐⭐⭐⭐（5/5）- API 注册非常简单
- **灵活性**：⭐⭐⭐⭐⭐（5/5）- 支持多种注册和废弃方式
- **可维护性**：⭐⭐⭐⭐（4/5）- 代码结构清晰，但缺少版本管理
- **性能**：⭐⭐⭐⭐（4/5）- 依赖数据库性能
- **安全性**：⭐⭐⭐（3/5）- 缺少权限控制和审计日志


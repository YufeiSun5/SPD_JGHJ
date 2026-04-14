# [前端规则] Vue 3 + Wails Desktop

## 长期事实

1. 当前前端使用 Vue 3 + Vite + Naive UI。
2. 路由采用 Hash 模式，以适配 Wails 桌面壳路径环境。
3. 页面主要通过 window.go.main.App 或 wailsjs 调用后端绑定方法。

## 修改前端时的默认规则

1. 优先复用现有 Wails 绑定能力，不要为了图省事先补 fetch/axios API 层。
2. 保持桌面语义，注意窗口控制、全屏、标题栏等 Wails 场景，不要按纯浏览器站点假设写交互。
3. 新增页面或复杂逻辑时，优先和现有页面风格、Naive UI 用法、目录结构保持一致。
4. 如果改动涉及业务核心页面或关键状态流，补充中英日三语说明。

## 禁止使用原生浏览器对话框

**严禁**在任何页面或组件中使用 `alert()`、`confirm()`、`prompt()`。  
项目已有两个统一的自研替代方案，必须优先使用：

### 1. `window.$message` — 轻量消息提示
- 由 `App.vue` 内的 `n-message-provider` + `src/components/MessageSetup.vue` 挂载到全局。
- 用于操作结果反馈（成功/失败/警告）：
  ```js
  window.$message.success('操作成功')
  window.$message.error('操作失败: ' + e)
  window.$message.warning('请先完成必填项')
  ```

### 2. `ConfirmDialog.vue` — 富信息确认弹窗
- 位于 `src/components/ConfirmDialog.vue`，支持 `type`（info/success/warning/danger）、`details` 详情列表、`warnings` 提示列表。
- 用于删除、结束班次等需要用户二次确认的危险操作，参考 `Production.vue` 中的用法。
- 使用 callback 模式（`onConfirm` / `onCancel`），不要用 Promise 包装。
- 引入与用法示例：
  ```vue
  import ConfirmDialog from '../components/ConfirmDialog.vue'

  const confirmDialog = ref({
    show: false, type: 'danger', title: '', message: '',
    details: [], warnings: [], warningTitle: '注意',
    confirmText: '确认', cancelText: '取消', confirmIcon: 'fas fa-check',
    onConfirm: () => {}, onCancel: () => { confirmDialog.value.show = false }
  })
  ```
  ```html
  <ConfirmDialog
    :show="confirmDialog.show"
    :type="confirmDialog.type"
    :title="confirmDialog.title"
    :message="confirmDialog.message"
    :details="confirmDialog.details"
    :warnings="confirmDialog.warnings"
    :warning-title="confirmDialog.warningTitle"
    :confirm-text="confirmDialog.confirmText"
    :cancel-text="confirmDialog.cancelText"
    :confirm-icon="confirmDialog.confirmIcon"
    @confirm="confirmDialog.onConfirm"
    @cancel="confirmDialog.onCancel"
  />
  ```

### 弹窗组件样式规范
- 蒙层：`background: rgba(0,0,0,0.7)`，无 `backdrop-filter`，无入场动画。
- 容器：`background: rgba(30,40,60,0.95)`，`backdrop-filter: blur(20px)`，`border-radius: 12px`。
- 头部：无渐变色背景，纯透明。
- 关闭按钮：`background: none`，无方块背景框，与 `TaskManagement.vue` 的 `close-btn` 风格一致。
- 参考文件：`src/views/TaskManagement.vue`（dialog 样式）、`src/components/ConfirmDialog.vue`（富确认框）。

## 改动后至少验证

1. wails dev 启动正常。
2. 页面能正确调用绑定方法。
3. 控制台没有 window.go 或 runtime 未定义类错误。

## 列表/表格数据的展示顺序约定

### ShiftReport.vue — 班次快照表格
- 排序由后端 `GetShiftSnapshots` 的 ORDER BY 保证，前端直接渲染 `snapshots` 数组，**不要在前端对数组重新排序**。
- 期望顺序：时间靠后的日期在上（`snapshot_date DESC`），同一天内夜班 → 中班 → 早班（`sys_shifts.sort_order DESC`），同班次内按设备升序。

### 筛选下拉框顺序
- **班次**（sys_shifts）：前端直接渲染后端返回顺序（sort_order ASC）。
- **班组**（sys_teams）：前端直接渲染后端返回顺序（id ASC），sys_teams 无 sort_order，**不要 reverse()**。
- **设备**、**人员**：同上，显示顺序以后端查询为准，前端仅做 a-z 排序（staffList 已按 name localeCompare）。

### 通用原则
- 若后端已在 SQL-level 保证排序，前端不做二次排序；若确实需要前端排序，必须基于真实字段（如 sort_order、created_at），不得用下标/遍历顺序逆转代替。

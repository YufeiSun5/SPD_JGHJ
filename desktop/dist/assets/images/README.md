# Logo 使用说明

## 当前Logo
- **文件**: `logo.svg`
- **用途**: 系统主Logo，显示在顶部导航栏和侧边栏
- **尺寸**: 
  - 顶部导航栏: 高度40px
  - 侧边栏: 高度35px
  - 宽度: 自适应

## 替换Logo

如果需要使用自己的Logo图片，请按以下步骤操作：

### 1. 准备Logo文件
- **推荐格式**: SVG（矢量图）或 PNG（透明背景）
- **推荐尺寸**: 
  - PNG: 至少 360x100 像素（宽x高）
  - SVG: 矢量图，任意尺寸
- **背景**: 透明背景最佳
- **颜色**: 建议使用白色或浅色文字/图形（因为背景是深色）

### 2. 替换文件
将您的Logo文件保存为以下任一名称：
- `logo.svg` （推荐，矢量图清晰度最佳）
- `logo.png` （位图，注意使用高分辨率）

放置在当前目录下：`desktop/frontend/public/assets/images/`

### 3. 如果使用PNG格式
需要修改以下两个文件中的引用路径：

**文件1**: `desktop/frontend/src/components/TopNav.vue`
```vue
<!-- 将 -->
<img src="/assets/images/logo.svg" alt="斯频德" class="logo-img">
<!-- 改为 -->
<img src="/assets/images/logo.png" alt="斯频德" class="logo-img">
```

**文件2**: `desktop/frontend/src/components/Sidebar.vue`
```vue
<!-- 将 -->
<img src="/assets/images/logo.svg" alt="斯频德" class="logo-img">
<!-- 改为 -->
<img src="/assets/images/logo.png" alt="斯频德" class="logo-img">
```

## 当前设计说明

当前的 `logo.svg` 是一个临时设计，包含：
- "斯频德" 中文文字（主标题）
- "焊机监测系统" 副标题
- "WELDING MONITOR" 英文副标题
- 工业风格装饰线条
- 配色: #546e7a 和 #607d8b（工业蓝灰色调）

您可以根据公司VI标准替换为正式Logo。


























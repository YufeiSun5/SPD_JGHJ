# Skill: SVG 旋转圆环中心校准

## 适用场景

当 SCADA 页面中的旋转圆环（SVG arc + 箭头）出现肉眼可见的"晃动"时，用本流程排查并校准旋转轴坐标。

---

## 原理说明

晃动有两类来源，需要区分处理：

| 类型 | 原因 | 是否可通过参数修复 |
|------|------|-------------------|
| **旋转轴偏移** | `transform-origin` 设置的坐标不在圆弧几何圆心上 | ✅ 可修复 |
| **箭头视觉不对称** | 箭头伸出圆弧范围，视觉重心随旋转公转 | ❌ 无法通过旋转轴参数消除，属正常现象 |

> **结论**：只需保证旋转轴在圆弧的几何圆心即可，箭头部分引起的视觉不对称不属于需要修复的问题。

---

## 第一步：量化当前晃动

在浏览器控制台执行，获取当前晃动半径：

```js
(function() {
  const el = document.getElementById('ringMask'); // 旋转元素 id
  const stage = document.getElementById('debugStage'); // 舞台容器 id
  const stageRect = stage.getBoundingClientRect();
  el.style.animationPlayState = 'paused';
  const duration = parseFloat(getComputedStyle(el).animationDuration) || 3;
  const angles = [0, 45, 90, 135, 180, 225, 270, 315];
  const pts = [];
  for (const deg of angles) {
    el.style.animationDelay = (-deg / 360 * duration) + 's';
    const b = el.getBoundingClientRect();
    pts.push({ cx: (b.left + b.right) / 2 - stageRect.left, cy: (b.top + b.bottom) / 2 - stageRect.top });
  }
  el.style.animationPlayState = '';
  el.style.animationDelay = '';
  const ax = pts.reduce((s, p) => s + p.cx, 0) / pts.length;
  const ay = pts.reduce((s, p) => s + p.cy, 0) / pts.length;
  const maxR = Math.max(...pts.map(p => Math.sqrt((p.cx - ax) ** 2 + (p.cy - ay) ** 2)));
  console.log('轨迹中心:', ax.toFixed(1), ay.toFixed(1), '| 晃动半径:', maxR.toFixed(1), 'px');
})();
```

---

## 第二步：计算圆弧真实圆心

在控制台执行，自动拟合圆弧几何圆心并输出 SVG 坐标：

```js
(function() {
  // 修改 selector 指向实际 <path> 元素
  const path = document.querySelector('#ringMask path');

  // 1. 采样圆弧上的点（路径本地坐标系）
  const total = path.getTotalLength();
  const pts = Array.from({ length: 16 }, (_, i) => {
    const p = path.getPointAtLength(i / 16 * total);
    return { x: p.x, y: p.y };
  });

  // 2. 最小二乘拟合圆心
  const n = pts.length;
  let sX=0,sY=0,sX2=0,sY2=0,sX3=0,sY3=0,sXY=0,sX2Y=0,sXY2=0;
  for (const p of pts) {
    sX+=p.x; sY+=p.y; sX2+=p.x*p.x; sY2+=p.y*p.y;
    sX3+=p.x**3; sY3+=p.y**3; sXY+=p.x*p.y;
    sX2Y+=p.x*p.x*p.y; sXY2+=p.x*p.y*p.y;
  }
  const A=2*(sX2-sX*sX/n), B=2*(sXY-sX*sY/n), C=sX3+sXY2-(sX2+sY2)*sX/n;
  const D=2*(sXY-sX*sY/n), E=2*(sY2-sY*sY/n), F=sX2Y+sY3-(sX2+sY2)*sY/n;
  const lcx=(C*E-F*B)/(A*E-D*B);
  const lcy=(A*F-D*C)/(A*E-D*B);

  // 3. 将路径本地坐标转换到 SVG viewport 坐标
  //    根据实际 <g transform="..."> 属性修改此处转换逻辑
  //    当前页面的变换链：translate(512,512) scale(10.24) scale(-1,1) translate(-50,-50)
  const svgX = (-(lcx - 50)) * 10.24 + 512;
  const svgY =   (lcy - 50)  * 10.24 + 512;

  console.log('路径本地圆心:', lcx.toFixed(3), lcy.toFixed(3));
  console.log('SVG 旋转轴坐标 → originX:', Math.round(svgX), ' originY:', Math.round(svgY));
})();
```

---

## 第三步：应用坐标

将上一步输出的 `originX / originY` 填入调试页或直接写入 CSS：

**调试页输入框：**
```
中心 X = <输出值>
中心 Y = <输出值>
```

**或直接写入 CSS `transform-origin`：**
```css
.ring-mask {
  transform-origin: 499px 512px; /* 从第二步计算所得 */
  animation: clockwise-spin 3s linear infinite;
}
```

---

## 注意事项

1. **路径本地坐标 → SVG 坐标的转换逻辑**依赖 `<g transform="...">` 属性，每个页面可能不同，需要对照实际 transform 链修改第二步代码中的换算部分。
2. 如果圆弧路径坐标是整数近似值（端点取整），拟合出的圆心会有 1~2px 偏差，这是正常的。
3. 箭头造成的 bbox 公转（通常 50px 量级）不是旋转轴问题，**不需要处理**。
4. 本流程仅适用于 CSS animation + `transform-origin` 方案；若使用 SMIL `animateTransform`，坐标解释方式不同。

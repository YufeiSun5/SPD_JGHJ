# 视频文件说明

## 文件位置
将监控视频文件放置在此目录下

## 文件命名规则
- `video1.mp4` - 第一个设备的监控视频（上方显示）
- `video2.mp4` - 第二个设备的监控视频（下方显示）

## 视频要求
- 格式：建议使用 MP4 (H.264) 格式
- 分辨率：建议 1280x720 或 1920x1080
- 文件大小：建议控制在 50MB 以内，避免加载过慢
- 注意：视频会自动循环播放且静音

## 使用方法

### 方法1：使用本地视频文件
1. 将视频文件重命名为 `video1.mp4` 和 `video2.mp4`
2. 放置到此目录
3. 在 `Cockpit.vue` 中取消注释视频路径：
   ```javascript
   videoSrc: `/src/assets/videos/video${index + 1}.mp4`
   ```

### 方法2：使用网络视频流
在 `Cockpit.vue` 中直接设置网络视频URL：
```javascript
videoSrc: 'http://your-video-stream-url'
```

## 当前状态
页面会根据后端 `GetAllDevices()` 接口返回的设备数量自动显示对应数量的视频监控区域（最多2个）。

























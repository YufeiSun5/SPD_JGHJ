// ============================================================================
// 定时扫描器 (Scanner) - 定时存储触发器
// ============================================================================
// 职责: 每1秒扫描内存Map, 检查定时存储周期
// 协程数: 1个单例
// 输出: CycleChan (定时存储, 缓冲500)
//
// 扫描逻辑:
//
//	for 每个Tag {
//	    if (store_mode=2或3) && (周期到期) {
//	        发送到 CycleChan
//	    }
//	}
//
// 为什么需要扫描器?
//
//	定时存储不依赖数据变化, 需要主动扫描
//	例如: 环境温度24小时不变, 但每小时要存一次
//
// 何时修改此文件:
//   - 需要调整扫描频率 (1秒 → 5秒)
//   - 需要添加新的扫描逻辑 (如定时报警检查)
//
// ============================================================================
package workers

import (
	"log"
	"time"

	"gin-mqtt-pgsql/core"
	"gin-mqtt-pgsql/models"
)

// StartScanner 启动定时扫描器 (x1) - 单例协程
func StartScanner() {
	log.Println("[Scanner] 启动定时扫描协程...")

	go scanner()

	log.Println("[Scanner] ✅ 定时扫描器已启动")
}

// scanner 扫描内存Map，检查定时存储
func scanner() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	log.Println("[Scanner] 开始扫描内存Map，检查定时存储点...")

	for now := range ticker.C {
		tagManager := core.GetTagManager()
		allTags := tagManager.GetAllTags()

		var scannedCount int
		var sentCount int

		for _, tag := range allTags {
			scannedCount++

			// 检查是否到达定时存储周期
			if tag.ShouldCycleStore(now) {
				// 根据数据类型获取值
				isString := (tag.DataType == "STRING" || tag.DataType == "TEXT")

				storeTask := &models.StoreTask{
					VarID:     tag.VarID,
					VarName:   tag.VarName,
					Value:     tag.GetValue(),
					StrValue:  tag.GetStringValue(),
					IsString:  isString,
					Timestamp: now,
				}

				// 非阻塞发送
				select {
				case core.CycleChan <- storeTask:
					sentCount++
					// 更新最后存储时间
					tag.UpdateStoreTime(now)
				default:
					log.Printf("[Scanner] CycleChan已满，跳过: %s", tag.VarName)
				}
			}
		}

		// if sentCount > 0 {
		// 	log.Printf("[Scanner] 扫描完成: 总数=%d, 发送=%d", scannedCount, sentCount)
		// }
	}
}

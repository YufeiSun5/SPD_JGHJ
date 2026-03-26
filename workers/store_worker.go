// ============================================================================
// 数据存储执行器 (Store Worker) - 批量写入优化
// ============================================================================
// 职责: 批量写入历史数据到数据库
// 协程数: ChangeWorker x5 + CycleWorker x5
// 输入: ChangeChan (变动存储, 缓冲500), CycleChan (定时存储, 缓冲500)
//
// 批量策略:
//
//	ChangeWorker: 每100ms 或 累积10条 → 批量写入
//	CycleWorker:  每200ms 或 累积20条 → 批量写入
//
// 为什么需要批量?
//
//	单条插入: 1000次/秒 = 1000次数据库连接 (性能瓶颈)
//	批量插入: 10次/秒 × 100条 = 1000条数据 (性能提升100倍)
//
// 何时修改此文件:
//   - 需要优化批量大小 (10条 → 50条)
//   - 需要调整写入频率 (100ms → 200ms)
//   - 需要修改存储格式 (添加新字段)
//
// ============================================================================
package workers

import (
	"log"
	"time"

	"gin-mqtt-pgsql/core"
	"gin-mqtt-pgsql/database"
	"gin-mqtt-pgsql/models"
)

// StartChangeWorkers 启动变动存储工作池 (x5) - IO密集
func StartChangeWorkers(count int) {
	log.Printf("[ChangeWorker] 启动 %d 个变动存储协程...", count)

	for i := 0; i < count; i++ {
		go changeWorker(i)
	}

	log.Printf("[ChangeWorker] ✅ 所有变动存储器已启动")
}

// changeWorker 单个变动存储协程
func changeWorker(id int) {
	log.Printf("[ChangeWorker-%d] 启动成功，等待任务...", id)

	// 批量写入优化: 每100ms或累积10条数据写一次
	batch := make([]*models.StoreTask, 0, 10)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case task, ok := <-core.ChangeChan:
			if !ok {
				// 通道关闭，写入剩余数据
				if len(batch) > 0 {
					flushBatch(id, batch)
				}
				log.Printf("[ChangeWorker-%d] 通道关闭，协程退出", id)
				return
			}

			batch = append(batch, task)

			// 达到批量大小，立即写入
			if len(batch) >= 10 {
				flushBatch(id, batch)
				batch = batch[:0] // 清空
			}

		case <-ticker.C:
			// 定时写入
			if len(batch) > 0 {
				flushBatch(id, batch)
				batch = batch[:0]
			}
		}
	}
}

// StartCycleWorkers 启动定时存储工作池 (x5) - IO密集
func StartCycleWorkers(count int) {
	log.Printf("[CycleWorker] 启动 %d 个定时存储协程...", count)

	for i := 0; i < count; i++ {
		go cycleWorker(i)
	}

	log.Printf("[CycleWorker] ✅ 所有定时存储器已启动")
}

// cycleWorker 单个定时存储协程
func cycleWorker(id int) {
	log.Printf("[CycleWorker-%d] 启动成功，等待任务...", id)

	batch := make([]*models.StoreTask, 0, 20)
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case task, ok := <-core.CycleChan:
			if !ok {
				if len(batch) > 0 {
					flushBatch(id, batch)
				}
				log.Printf("[CycleWorker-%d] 通道关闭，协程退出", id)
				return
			}

			batch = append(batch, task)

			if len(batch) >= 20 {
				flushBatch(id, batch)
				batch = batch[:0]
			}

		case <-ticker.C:
			if len(batch) > 0 {
				flushBatch(id, batch)
				batch = batch[:0]
			}
		}
	}
}

// flushBatch 批量写入数据库
func flushBatch(workerID int, batch []*models.StoreTask) {
	if err := database.InsertHistoryDataBatch(batch); err != nil {
		log.Printf("[Worker-%d] 批量写入失败 (%d条): %v", workerID, len(batch), err)
	}
	// else {
	// 	log.Printf("[Worker-%d] ✅ 批量写入成功 (%d条)", workerID, len(batch))
	// }
}

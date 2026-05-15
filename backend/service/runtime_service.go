package service

import (
	"fmt"
	"time"

	"gin-mqtt-pgsql/core"
	"gin-mqtt-pgsql/workers"
)

// TagData is the realtime tag DTO shared by Wails and future local web transports.
// CN: Wails 和后续本机 Web 壳共用的实时测点传输结构，业务数据来自内存 TagManager。
// EN: Shared realtime tag DTO for Wails and future local web transports; data comes from the in-memory TagManager.
// JP: Wails と将来のローカル Web トランスポートで共用するリアルタイム点位 DTO。データ源はメモリ上の TagManager。
type TagData struct {
	VarName     string `json:"var_name"`
	DisplayName string `json:"display_name"`
	DataType    string `json:"data_type"`
	Value       string `json:"value"`
	Unit        string `json:"unit"`
	AlarmState  string `json:"alarm_state"`
}

// TagInfo is the compact tag metadata DTO used by history and selection screens.
// CN: 历史查询和下拉选择使用的精简点位元数据，避免页面直接依赖完整 Tag 内部状态。
// EN: Compact tag metadata for history queries and pickers, keeping UI away from full Tag internals.
// JP: 履歴検索と選択 UI 用の簡易点位メタデータ。完全な Tag 内部状態へ依存させない。
type TagInfo struct {
	VarID       int64  `json:"var_id"`
	VarName     string `json:"var_name"`
	DisplayName string `json:"display_name"`
	Unit        string `json:"unit"`
	StoreMode   int    `json:"store_mode"`
	DataType    string `json:"data_type"`
}

func GetRealtimeData() []TagData {
	tagManager := core.GetTagManager()
	allTags := tagManager.GetAllTags()

	result := make([]TagData, 0, len(allTags))
	for _, tag := range allTags {
		currentValue := tag.GetValue()
		currentStrValue := tag.GetStringValue()
		alarmState, _ := tag.GetAlarmState()

		data := TagData{
			VarName:     tag.VarName,
			DisplayName: tag.DisplayName,
			DataType:    tag.DataType,
			Unit:        tag.Unit,
			AlarmState:  alarmState,
		}

		if tag.DataType == "STRING" || tag.DataType == "TEXT" {
			if currentStrValue == "" {
				data.Value = "-"
			} else {
				data.Value = currentStrValue
			}
		} else if currentValue == float64(int64(currentValue)) {
			data.Value = fmt.Sprintf("%d", int64(currentValue))
		} else {
			data.Value = fmt.Sprintf("%.2f", currentValue)
		}

		result = append(result, data)
	}

	return result
}

func GetAllTags() []TagInfo {
	tagManager := core.GetTagManager()
	allTags := tagManager.GetAllTags()

	result := make([]TagInfo, 0, len(allTags))
	for _, tag := range allTags {
		result = append(result, TagInfo{
			VarID:       tag.VarID,
			VarName:     tag.VarName,
			DisplayName: tag.DisplayName,
			Unit:        tag.Unit,
			StoreMode:   tag.StoreMode,
			DataType:    tag.DataType,
		})
	}

	return result
}

func GetSystemMonitor() map[string]interface{} {
	channelStats := core.GetChannelStats()

	scheduler := workers.GetTaskScheduler()
	var taskStats map[string]int
	if scheduler != nil {
		taskStats = scheduler.GetTaskCount()
	} else {
		taskStats = map[string]int{
			"scheduled":   0,
			"data_change": 0,
			"condition":   0,
		}
	}

	channelCapacity := map[string]int{
		"logic_chan":   2000,
		"change_chan":  500,
		"cycle_chan":   500,
		"alarm_chan":   200,
		"sse_chan":     100,
		"event_chan":   300,
		"trigger_chan": 10000,
	}

	channelUsage := make(map[string]float64)
	for name, current := range channelStats {
		capacity := channelCapacity[name]
		if capacity > 0 {
			channelUsage[name] = float64(current) / float64(capacity) * 100
		}
	}

	return map[string]interface{}{
		"channel_stats":    channelStats,
		"channel_capacity": channelCapacity,
		"channel_usage":    channelUsage,
		"task_stats":       taskStats,
		"alerts":           core.CheckChannelHealth(),
		"timestamp":        time.Now(),
	}
}

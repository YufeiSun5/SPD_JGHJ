package service

import (
	"encoding/json"
	"fmt"
	"time"

	"gin-mqtt-pgsql/database"
	"gin-mqtt-pgsql/models"
)

func GetAllOrders() ([]*models.ProOrder, error) {
	return database.GetAllOrders(nil, nil)
}

func CreateOrder(orderNo, productCode string, planQty int, targetDeviceID *int) (*models.ProOrder, error) {
	order := &models.ProOrder{
		OrderNo:        orderNo,
		ProductCode:    productCode,
		TargetDeviceID: targetDeviceID,
		PlanQty:        planQty,
		ActualQty:      0,
		OkQty:          0,
		NgQty:          0,
		Status:         0,
		Version:        0,
	}

	if err := database.CreateOrder(order); err != nil {
		return nil, err
	}

	return order, nil
}

// UpdateOrder keeps the production-order state machine outside the Wails binding.
// CN: 工单状态流转在 service 中统一处理，Wails/Web 壳只传入请求参数，避免多个入口复制状态机。
// EN: The order state machine lives in service code so Wails/Web transports only pass request parameters.
// JP: 工単状態遷移は service に集約し、Wails/Web 側はリクエスト引数を渡すだけにする。
func UpdateOrder(id int64, productCode *string, planQty *int, status *int8, targetDeviceID *int) error {
	updates := make(map[string]interface{})

	if productCode != nil {
		updates["product_code"] = *productCode
	}
	if planQty != nil {
		updates["plan_qty"] = *planQty
	}
	if targetDeviceID != nil {
		updates["target_device_id"] = *targetDeviceID
	}
	if status != nil {
		order, err := database.GetOrderByID(id)
		if err != nil {
			return err
		}

		updates["status"] = *status

		now := time.Now()
		if *status == 1 {
			updates["current_start_time"] = now
			if order.StartTime == nil {
				updates["start_time"] = now
			}
		}
		if *status == 2 && order.Status == 1 && order.CurrentStartTime != nil {
			elapsed := int(time.Since(*order.CurrentStartTime).Seconds())
			updates["used_seconds"] = order.UsedSeconds + elapsed
			updates["current_start_time"] = nil
		}
		if *status == 3 {
			if order.CurrentStartTime != nil {
				elapsed := int(time.Since(*order.CurrentStartTime).Seconds())
				updates["used_seconds"] = order.UsedSeconds + elapsed
			}
			updates["end_time"] = now
			updates["current_start_time"] = nil
		}
	}

	return database.UpdateOrder(id, updates)
}

func DeleteOrder(id int64) error {
	return database.DeleteOrder(id)
}

func StartProductionSmart(orderID int64) (*models.ProProductionRun, error) {
	order, err := database.GetOrderByID(orderID)
	if err != nil {
		return nil, fmt.Errorf("工单不存在: %w", err)
	}
	if order.TargetDeviceID == nil {
		return nil, fmt.Errorf("工单未指定设备")
	}

	deviceID := *order.TargetDeviceID
	session, err := database.GetActiveSession(deviceID)
	if err != nil {
		return nil, fmt.Errorf("设备%d没有活动班次，请先在\"人员管理\"页面进行班次登记", deviceID)
	}

	var staffIDs []int
	if err := json.Unmarshal([]byte(session.StaffIDs), &staffIDs); err != nil {
		return nil, fmt.Errorf("解析班次员工信息失败: %w", err)
	}

	req := &models.StartProductionRequest{
		OrderID:     orderID,
		DeviceID:    deviceID,
		TeamID:      session.TeamID,
		OperatorIDs: staffIDs,
		Remark:      nil,
	}

	return database.StartProductionSmart(req)
}

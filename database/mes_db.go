package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"gin-mqtt-pgsql/models"

	"gorm.io/gorm"
)

// ========================================================
// 班组管理 (sys_teams)
// ========================================================

// CreateTeam 创建班组
func CreateTeam(team *models.SysTeam) error {
	result := DB.Create(team)
	if result.Error != nil {
		return fmt.Errorf("创建班组失败: %w", result.Error)
	}
	return nil
}

// GetTeamByID 根据ID获取班组
func GetTeamByID(id int) (*models.SysTeam, error) {
	var team models.SysTeam
	result := DB.First(&team, id)
	if result.Error != nil {
		return nil, fmt.Errorf("查询班组失败: %w", result.Error)
	}
	return &team, nil
}

// GetAllTeams 获取所有班组
func GetAllTeams(status *int8) ([]*models.SysTeam, error) {
	var teams []*models.SysTeam
	query := DB.Model(&models.SysTeam{})

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	result := query.Find(&teams)
	if result.Error != nil {
		return nil, fmt.Errorf("查询班组列表失败: %w", result.Error)
	}
	return teams, nil
}

// UpdateTeam 更新班组
func UpdateTeam(id int, updates map[string]interface{}) error {
	result := DB.Model(&models.SysTeam{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("更新班组失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("班组不存在: id=%d", id)
	}
	return nil
}

// DeleteTeam 删除班组
func DeleteTeam(id int) error {
	result := DB.Delete(&models.SysTeam{}, id)
	if result.Error != nil {
		return fmt.Errorf("删除班组失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("班组不存在: id=%d", id)
	}
	return nil
}

// ========================================================
// 人员管理 (sys_staff)
// ========================================================

// CreateStaff 创建员工
func CreateStaff(staff *models.SysStaff) error {
	result := DB.Create(staff)
	if result.Error != nil {
		return fmt.Errorf("创建员工失败: %w", result.Error)
	}
	return nil
}

// GetStaffByID 根据ID获取员工（包含班组信息）
func GetStaffByID(id int) (*models.SysStaff, error) {
	var staff models.SysStaff
	result := DB.Preload("CurrentTeam").First(&staff, id)
	if result.Error != nil {
		return nil, fmt.Errorf("查询员工失败: %w", result.Error)
	}
	return &staff, nil
}

// GetStaffByCode 根据工号获取员工
func GetStaffByCode(staffCode string) (*models.SysStaff, error) {
	var staff models.SysStaff
	result := DB.Preload("CurrentTeam").Where("staff_code = ?", staffCode).First(&staff)
	if result.Error != nil {
		return nil, fmt.Errorf("查询员工失败: %w", result.Error)
	}
	return &staff, nil
}

// GetAllStaff 获取所有员工（支持筛选）
func GetAllStaff(teamID *int, isActive *int8) ([]*models.SysStaff, error) {
	var staffList []*models.SysStaff
	query := DB.Model(&models.SysStaff{}).Preload("CurrentTeam")

	if teamID != nil {
		query = query.Where("current_team_id = ?", *teamID)
	}
	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	result := query.Find(&staffList)
	if result.Error != nil {
		return nil, fmt.Errorf("查询员工列表失败: %w", result.Error)
	}
	return staffList, nil
}

// UpdateStaff 更新员工信息
func UpdateStaff(id int, updates map[string]interface{}) error {
	fmt.Printf("🔧 [UpdateStaff] 开始更新员工 id=%d, updates=%+v\n", id, updates)

	// 如果更新了班组，需要记录调动历史
	if newTeamID, ok := updates["current_team_id"]; ok {
		fmt.Printf("📋 [UpdateStaff] 检测到班组更新: newTeamID=%v\n", newTeamID)

		// 1. 获取员工当前的班组
		var staff models.SysStaff
		fmt.Printf("🔍 [UpdateStaff] 正在查询员工 id=%d...\n", id)

		if err := DB.Model(&models.SysStaff{}).Where("id = ?", id).First(&staff).Error; err != nil {
			fmt.Printf("❌ [UpdateStaff] 查询员工失败: id=%d, error=%v\n", id, err)

			// 尝试直接查询看看员工是否存在
			var count int64
			DB.Model(&models.SysStaff{}).Where("id = ?", id).Count(&count)
			fmt.Printf("🔍 [UpdateStaff] 直接计数查询结果: count=%d\n", count)

			return fmt.Errorf("员工不存在: id=%d, %w", id, err)
		}

		fmt.Printf("✅ [UpdateStaff] 查询到员工: id=%d, name=%s, current_team_id=%v\n", staff.ID, staff.Name, staff.CurrentTeamID)

		// 2. 如果班组确实发生了变化
		if newTeamID != nil && (staff.CurrentTeamID == nil || *staff.CurrentTeamID != newTeamID.(int)) {
			// 使用事务确保数据一致性
			tx := DB.Begin()
			defer func() {
				if r := recover(); r != nil {
					tx.Rollback()
				}
			}()

			// 3. 如果有旧班组，记录"离开"
			if staff.CurrentTeamID != nil {
				history := &models.SysStaffHistory{
					StaffID:    id,
					TeamID:     *staff.CurrentTeamID,
					ActionType: 2, // 离开班组
				}
				if err := tx.Create(history).Error; err != nil {
					tx.Rollback()
					return fmt.Errorf("记录离开班组失败: %w", err)
				}
			}

			// 4. 记录"加入"新班组
			history := &models.SysStaffHistory{
				StaffID:    id,
				TeamID:     newTeamID.(int),
				ActionType: 1, // 加入班组
			}
			if err := tx.Create(history).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("记录加入班组失败: %w", err)
			}

			// 5. 更新员工信息
			if err := tx.Model(&models.SysStaff{}).Where("id = ?", id).Updates(updates).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("更新员工失败: %w", err)
			}

			// 6. 提交事务
			if err := tx.Commit().Error; err != nil {
				return fmt.Errorf("提交事务失败: %w", err)
			}

			return nil
		}
	}

	// 普通更新（不涉及班组变更）
	result := DB.Model(&models.SysStaff{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("更新员工失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("员工不存在: id=%d", id)
	}
	return nil
}

// DeleteStaff 删除员工
func DeleteStaff(id int) error {
	result := DB.Delete(&models.SysStaff{}, id)
	if result.Error != nil {
		return fmt.Errorf("删除员工失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("员工不存在: id=%d", id)
	}
	return nil
}

// TransferStaff 调动员工（包含历史记录）
func TransferStaff(req *models.TransferStaffRequest) error {
	// 开启事务
	tx := DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 获取当前员工信息
	var staff models.SysStaff
	if err := tx.First(&staff, req.StaffID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("员工不存在: %w", err)
	}

	// 2. 如果原来有班组，记录离开历史
	if staff.CurrentTeamID != nil {
		leaveHistory := &models.SysStaffHistory{
			StaffID:      req.StaffID,
			TeamID:       *staff.CurrentTeamID,
			ActionType:   2, // 离开
			HappenedAt:   time.Now(),
			OperatorName: req.OperatorName,
		}
		if err := tx.Create(leaveHistory).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("记录离开历史失败: %w", err)
		}
	}

	// 3. 记录加入新班组历史
	joinHistory := &models.SysStaffHistory{
		StaffID:      req.StaffID,
		TeamID:       req.NewTeamID,
		ActionType:   1, // 加入
		HappenedAt:   time.Now(),
		OperatorName: req.OperatorName,
	}
	if err := tx.Create(joinHistory).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("记录加入历史失败: %w", err)
	}

	// 4. 更新员工当前班组
	if err := tx.Model(&staff).Update("current_team_id", req.NewTeamID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新员工班组失败: %w", err)
	}

	// 提交事务
	return tx.Commit().Error
}

// GetStaffHistory 获取员工调动历史
func GetStaffHistory(staffID int) ([]*models.SysStaffHistory, error) {
	var history []*models.SysStaffHistory
	result := DB.Preload("Team").Where("staff_id = ?", staffID).Order("happened_at DESC").Find(&history)
	if result.Error != nil {
		return nil, fmt.Errorf("查询员工历史失败: %w", result.Error)
	}
	return history, nil
}

// GetTeamHistoryAtTime 获取某班组在某时刻的人员名单
func GetTeamHistoryAtTime(teamID int, targetTime time.Time) ([]*models.SysStaff, error) {
	// 复杂查询: 找出在targetTime时刻属于该班组的所有员工
	// 逻辑: 最后一次加入该班组的时间 < targetTime && (没有离开记录 OR 离开时间 > targetTime)
	var staffIDs []int

	query := `
		SELECT DISTINCT staff_id 
		FROM sys_staff_history h1
		WHERE team_id = ? 
		  AND happened_at <= ?
		  AND action_type = 1
		  AND NOT EXISTS (
			SELECT 1 FROM sys_staff_history h2 
			WHERE h2.staff_id = h1.staff_id 
			  AND h2.team_id = ? 
			  AND h2.action_type = 2 
			  AND h2.happened_at > h1.happened_at 
			  AND h2.happened_at <= ?
		  )
	`

	if err := DB.Raw(query, teamID, targetTime, teamID, targetTime).Scan(&staffIDs).Error; err != nil {
		return nil, fmt.Errorf("查询历史人员失败: %w", err)
	}

	if len(staffIDs) == 0 {
		return []*models.SysStaff{}, nil
	}

	var staffList []*models.SysStaff
	if err := DB.Where("id IN ?", staffIDs).Find(&staffList).Error; err != nil {
		return nil, fmt.Errorf("查询员工信息失败: %w", err)
	}

	return staffList, nil
}

// ========================================================
// 工单管理 (pro_orders)
// ========================================================

// CreateOrder 创建工单
func CreateOrder(order *models.ProOrder) error {
	result := DB.Create(order)
	if result.Error != nil {
		return fmt.Errorf("创建工单失败: %w", result.Error)
	}
	return nil
}

// GetOrderByID 根据ID获取工单
func GetOrderByID(id int64) (*models.ProOrder, error) {
	var order models.ProOrder
	result := DB.First(&order, id)
	if result.Error != nil {
		return nil, fmt.Errorf("查询工单失败: %w", result.Error)
	}
	return &order, nil
}

// GetOrderByNo 根据工单号获取工单
func GetOrderByNo(orderNo string) (*models.ProOrder, error) {
	var order models.ProOrder
	result := DB.Where("order_no = ?", orderNo).First(&order)
	if result.Error != nil {
		return nil, fmt.Errorf("查询工单失败: %w", result.Error)
	}
	return &order, nil
}

// GetAllOrders 获取所有工单（支持筛选）
func GetAllOrders(status *int8, deviceID *int) ([]*models.ProOrder, error) {
	var orders []*models.ProOrder
	query := DB.Model(&models.ProOrder{})

	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if deviceID != nil {
		query = query.Where("target_device_id = ?", *deviceID)
	}

	result := query.Order("created_at DESC").Find(&orders)
	if result.Error != nil {
		return nil, fmt.Errorf("查询工单列表失败: %w", result.Error)
	}
	return orders, nil
}

// GetLastNonzeroHistoryValue 返回指定变量在触发时间之前的最近非零历史值。
// CN: 产量任务用它识别计数器重连恢复，避免把 old=0,new=历史累计值误写成新产量。
// EN: Production tasks use this to detect counter reconnect recovery before writing order totals.
// JP: 産量タスクはこれでカウンタ復旧値を検出し、累積値を新規産量として誤記録しない。
func GetLastNonzeroHistoryValue(varID int64, before time.Time) (int, bool, error) {
	var row struct {
		Val float64 `gorm:"column:val"`
	}
	err := DB.Table("sys_data_history").
		Select("val").
		Where("var_id = ? AND created_at < ? AND val > 0", varID, before).
		Order("created_at DESC").
		Limit(1).
		Scan(&row).Error
	if err != nil {
		return 0, false, err
	}
	if row.Val <= 0 {
		return 0, false, nil
	}
	return int(row.Val), true, nil
}

// UpdateOrder 更新工单
func UpdateOrder(id int64, updates map[string]interface{}) error {
	// 先检查记录是否存在
	var existingOrder models.ProOrder
	if err := DB.First(&existingOrder, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("工单不存在: id=%d", id)
		}
		return fmt.Errorf("查询工单失败: %w", err)
	}

	// 过滤掉空值和相同值的更新
	filteredUpdates := make(map[string]interface{})
	for key, newValue := range updates {
		if newValue == nil {
			continue // 跳过空值
		}

		// 检查值是否真的有变化
		var currentValue interface{}
		switch key {
		case "product_code":
			currentValue = existingOrder.ProductCode
		case "plan_qty":
			currentValue = existingOrder.PlanQty
		case "status":
			currentValue = existingOrder.Status
		case "target_device_id":
			currentValue = existingOrder.TargetDeviceID
			if newDeviceID, ok := newValue.(int); ok {
				currentID := 0
				if existingOrder.TargetDeviceID != nil {
					currentID = *existingOrder.TargetDeviceID
				}
				if currentID != newDeviceID {
					var runCount int64
					if err := DB.Model(&models.ProProductionRun{}).
						Where("order_id = ? AND (end_time IS NULL OR run_ok_qty <> 0 OR run_ng_qty <> 0)", id).
						Count(&runCount).Error; err != nil {
						return fmt.Errorf("检查工单运行记录失败: %w", err)
					}
					if runCount > 0 {
						return fmt.Errorf("工单已有生产运行记录，禁止修改目标设备")
					}
				}
			}
		case "start_time":
			currentValue = existingOrder.StartTime
		case "end_time":
			currentValue = existingOrder.EndTime
		default:
			// 未知字段，直接加入更新
			filteredUpdates[key] = newValue
			continue
		}

		// 只有值真的不同时才加入更新
		if currentValue != newValue {
			filteredUpdates[key] = newValue
		}
	}

	// 如果没有需要更新的字段，直接返回成功
	if len(filteredUpdates) == 0 {
		fmt.Printf("ℹ️ 工单无需更新 - ID: %d, 所有值都相同\n", id)
		return nil
	}

	isCompleting := false
	if status, ok := filteredUpdates["status"]; ok {
		switch v := status.(type) {
		case int8:
			isCompleting = v == 3
		case int:
			isCompleting = v == 3
		}
	}

	updateFn := func(tx *gorm.DB) error {
		result := tx.Model(&models.ProOrder{}).Where("id = ?", id).Updates(filteredUpdates)
		if result.Error != nil {
			return fmt.Errorf("更新工单失败: %w", result.Error)
		}

		if isCompleting {
			endTime, ok := filteredUpdates["end_time"].(time.Time)
			if !ok {
				endTime = time.Now()
			}
			// CN: 工单完工必须同时关闭活动运行记录，否则后续产量会继续写入旧工单。
			// EN: Completing an order must close active runs, or later output can still hit the old order.
			// JP: 工単完了時は活動中の実行記録も閉じ、以後の産量が旧工単へ入らないようにする。
			if err := tx.Model(&models.ProProductionRun{}).
				Where("order_id = ? AND end_time IS NULL", id).
				Update("end_time", endTime).Error; err != nil {
				return fmt.Errorf("关闭工单运行记录失败: %w", err)
			}
		}
		fmt.Printf("✅ 工单更新成功 - ID: %d, 影响行数: %d, 更新字段: %+v\n", id, result.RowsAffected, filteredUpdates)
		return nil
	}

	if isCompleting {
		if err := DB.Transaction(updateFn); err != nil {
			return err
		}
	} else if err := updateFn(DB); err != nil {
		return err
	}

	return nil
}

// DeleteOrder 删除工单
func DeleteOrder(id int64) error {
	result := DB.Delete(&models.ProOrder{}, id)
	if result.Error != nil {
		return fmt.Errorf("删除工单失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("工单不存在: id=%d", id)
	}
	return nil
}

// UpdateOrderQuantities 更新工单数量（带乐观锁）
func UpdateOrderQuantities(orderID int64, okQtyDelta, ngQtyDelta int) error {
	// 乐观锁更新
	result := DB.Model(&models.ProOrder{}).
		Where("id = ?", orderID).
		Updates(map[string]interface{}{
			"actual_qty": DB.Raw("actual_qty + ?", okQtyDelta+ngQtyDelta),
			"ok_qty":     DB.Raw("ok_qty + ?", okQtyDelta),
			"ng_qty":     DB.Raw("ng_qty + ?", ngQtyDelta),
			"version":    DB.Raw("version + 1"),
		})

	if result.Error != nil {
		return fmt.Errorf("更新工单数量失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("工单不存在或版本冲突: id=%d", orderID)
	}
	return nil
}

// GetOrderSummary 获取工单汇总信息（包含运行记录）
func GetOrderSummary(orderID int64) (*models.OrderSummaryResponse, error) {
	// 1. 获取工单基础信息
	var order models.ProOrder
	if err := DB.First(&order, orderID).Error; err != nil {
		return nil, fmt.Errorf("查询工单失败: %w", err)
	}

	// 2. 获取所有运行记录
	var runs []models.ProProductionRun
	if err := DB.Preload("Team").Where("order_id = ?", orderID).Find(&runs).Error; err != nil {
		return nil, fmt.Errorf("查询运行记录失败: %w", err)
	}

	// 3. 计算统计数据
	var totalDuration float64
	for _, run := range runs {
		if run.EndTime != nil {
			duration := run.EndTime.Sub(run.StartTime).Hours()
			totalDuration += duration
		}
	}

	avgEfficiency := 0.0
	if order.PlanQty > 0 {
		avgEfficiency = float64(order.ActualQty) / float64(order.PlanQty) * 100
	}

	summary := &models.OrderSummaryResponse{
		ProOrder:      order,
		RunCount:      len(runs),
		TotalDuration: totalDuration,
		AvgEfficiency: avgEfficiency,
		Runs:          runs,
	}

	return summary, nil
}

// ========================================================
// 生产运行记录管理 (pro_production_runs)
// ========================================================

// StartProduction 开始生产（创建运行记录）
func StartProduction(req *models.StartProductionRequest) (*models.ProProductionRun, error) {
	// 开启事务
	tx := DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 验证工单存在且状态正确
	var order models.ProOrder
	if err := tx.First(&order, req.OrderID).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("工单不存在: %w", err)
	}

	// 2. 将操作员ID数组转为JSON字符串
	operatorIDsJSON, err := json.Marshal(req.OperatorIDs)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("操作员ID序列化失败: %w", err)
	}

	// 3. 创建运行记录
	run := &models.ProProductionRun{
		OrderID:     req.OrderID,
		DeviceID:    req.DeviceID,
		TeamID:      req.TeamID,
		RunOkQty:    0,
		RunNgQty:    0,
		StartTime:   time.Now(),
		EndTime:     nil,
		OperatorIDs: string(operatorIDsJSON),
		Remark:      req.Remark,
	}

	if err := tx.Create(run).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("创建运行记录失败: %w", err)
	}

	// 4. 更新工单状态为"生产中"
	now := time.Now()
	updates := map[string]interface{}{
		"status": 1, // 生产中
	}
	// 如果是首次开工，记录开工时间
	if order.StartTime == nil {
		updates["start_time"] = now
	}

	if err := tx.Model(&models.ProOrder{}).Where("id = ?", req.OrderID).Updates(updates).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("更新工单状态失败: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	// 重新加载包含关联数据
	DB.Preload("Team").First(run, run.ID)
	return run, nil
}

// GetProductionRunByID 根据ID获取运行记录
func GetProductionRunByID(id int64) (*models.ProProductionRun, error) {
	var run models.ProProductionRun
	result := DB.Preload("Order").Preload("Team").First(&run, id)
	if result.Error != nil {
		return nil, fmt.Errorf("查询运行记录失败: %w", result.Error)
	}
	return &run, nil
}

// GetActiveRun 获取某设备当前正在运行的记录
func GetActiveRun(deviceID int) (*models.ProProductionRun, error) {
	var run models.ProProductionRun
	result := DB.Preload("Order").Preload("Team").
		Where("device_id = ? AND end_time IS NULL", deviceID).
		First(&run)
	if result.Error != nil {
		return nil, fmt.Errorf("查询活动运行记录失败: %w", result.Error)
	}
	return &run, nil
}

// UpdateProductionRun 更新运行记录产量
func UpdateProductionRun(runID int64, okQtyDelta, ngQtyDelta int) error {
	// 开启事务
	tx := DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 获取运行记录并确认关联工单仍处于生产中。
	var run models.ProProductionRun
	if err := tx.First(&run, runID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("查询运行记录失败: %w", err)
	}

	var currentOrder models.ProOrder
	if err := tx.First(&currentOrder, run.OrderID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("查询工单失败: %w", err)
	}
	if run.EndTime != nil || currentOrder.Status != 1 {
		tx.Rollback()
		return fmt.Errorf("运行记录未处于生产中，禁止更新产量: run_id=%d, order_id=%d, order_status=%d", run.ID, run.OrderID, currentOrder.Status)
	}

	// 2. 更新运行记录
	result := tx.Model(&models.ProProductionRun{}).
		Where("id = ? AND end_time IS NULL", runID).
		Updates(map[string]interface{}{
			"run_ok_qty": DB.Raw("run_ok_qty + ?", okQtyDelta),
			"run_ng_qty": DB.Raw("run_ng_qty + ?", ngQtyDelta),
		})

	if result.Error != nil {
		tx.Rollback()
		return fmt.Errorf("更新运行记录失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		tx.Rollback()
		return fmt.Errorf("运行记录不存在: id=%d", runID)
	}

	// 3. 更新工单总数量
	// 🔥 业务逻辑说明:
	//   - actual_qty (总产量) = 只在良品增加时增长 (okQtyDelta > 0)
	//   - 如果 okQtyDelta < 0 (NG转换), actual_qty 不变
	//   - ok_qty 和 ng_qty 按增量更新
	actualQtyDelta := okQtyDelta + ngQtyDelta
	if okQtyDelta < 0 && ngQtyDelta > 0 {
		// NG转换场景: ok_qty -1, ng_qty +1, actual_qty 不变
		actualQtyDelta = 0
	}

	// 检查良品数量是否足够扣减
	if okQtyDelta < 0 && currentOrder.OkQty+okQtyDelta < 0 {
		tx.Rollback()
		return fmt.Errorf("良品数量不足，无法扣减 %d (当前良品数: %d)", -okQtyDelta, currentOrder.OkQty)
	}

	if err := tx.Model(&models.ProOrder{}).
		Where("id = ?", run.OrderID).
		Updates(map[string]interface{}{
			"actual_qty": DB.Raw("actual_qty + ?", actualQtyDelta),
			"ok_qty":     DB.Raw("ok_qty + ?", okQtyDelta),
			"ng_qty":     DB.Raw("ng_qty + ?", ngQtyDelta),
		}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新工单数量失败: %w", err)
	}

	// 提交事务
	return tx.Commit().Error
}

// EndProduction 结束生产（关闭运行记录）
func EndProduction(runID int64, remark string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"end_time": now,
	}
	if remark != "" {
		updates["remark"] = remark
	}

	result := DB.Model(&models.ProProductionRun{}).
		Where("id = ? AND end_time IS NULL", runID).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("结束生产失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("运行记录不存在或已结束: id=%d", runID)
	}
	return nil
}

// GetOrderRuns 获取某工单的所有运行记录
func GetOrderRuns(orderID int64) ([]*models.ProProductionRun, error) {
	var runs []*models.ProProductionRun
	result := DB.Preload("Team").
		Where("order_id = ?", orderID).
		Order("start_time DESC").
		Find(&runs)
	if result.Error != nil {
		return nil, fmt.Errorf("查询运行记录失败: %w", result.Error)
	}
	return runs, nil
}

// GetDeviceRuns 获取某设备的运行历史
func GetDeviceRuns(deviceID int, startTime, endTime *time.Time) ([]*models.ProProductionRun, error) {
	var runs []*models.ProProductionRun
	query := DB.Preload("Order").Preload("Team").Where("device_id = ?", deviceID)

	// 时间范围筛选 - 查询与时间范围有重叠的记录
	if startTime != nil && endTime != nil {
		// 记录的结束时间 > 查询开始时间 AND 记录的开始时间 <= 查询结束时间
		// 使用 <= 而不是 < 以包含在结束时间点开始的记录
		query = query.Where("(end_time IS NULL OR end_time > ?) AND start_time <= ?", *startTime, *endTime)
	} else if startTime != nil {
		// 只有开始时间：记录的结束时间 > 查询开始时间 OR 记录还未结束
		query = query.Where("end_time IS NULL OR end_time > ?", *startTime)
	} else if endTime != nil {
		// 只有结束时间：记录的开始时间 <= 查询结束时间
		query = query.Where("start_time <= ?", *endTime)
	}

	result := query.Order("start_time DESC").Find(&runs)
	if result.Error != nil {
		return nil, fmt.Errorf("查询设备运行历史失败: %w", result.Error)
	}
	return runs, nil
}

// IncrementProductionQtyByDevice 根据设备ID增加产量（同时更新工单和班次）
func IncrementProductionQtyByDevice(deviceID int, okQtyDelta, ngQtyDelta int) error {
	// 1. 查找该设备当前活动且工单仍处于生产中的运行记录。
	// CN: 只允许生产中工单接收产量，避免完工/暂停/关闭工单继续吃数。
	// EN: Only running orders may receive output, preventing completed/paused/closed orders from being updated.
	// JP: 生産中の工単のみ産量を受け取り、完了・一時停止・終了工単への誤加算を防ぐ。
	var run models.ProProductionRun
	err := DB.Joins("JOIN pro_orders o ON o.id = pro_production_runs.order_id").
		Where("pro_production_runs.device_id = ? AND pro_production_runs.end_time IS NULL AND o.status = ?", deviceID, 1).
		Order("pro_production_runs.id DESC").
		First(&run).Error

	if err != nil {
		return fmt.Errorf("设备%d没有生产中的活动运行记录，请先在生产管理页面点击\"开工\"", deviceID)
	}

	// 2. 调用现有的 UpdateProductionRun 方法（会自动更新工单）
	return UpdateProductionRun(run.ID, okQtyDelta, ngQtyDelta)
}

// StartProductionSmart 智能开工（自动暂停该设备的其他工单）
func StartProductionSmart(req *models.StartProductionRequest) (*models.ProProductionRun, error) {
	// 开启事务
	tx := DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 查找该设备当前所有"生产中"的工单，暂停它们
	var runningOrders []models.ProOrder
	if err := tx.Where("target_device_id = ? AND status = 1", req.DeviceID).
		Find(&runningOrders).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("查询设备工单失败: %w", err)
	}

	// 暂停所有正在生产的工单
	for _, order := range runningOrders {
		if order.ID != req.OrderID { // 不是当前要开工的工单
			log.Printf("[Database] 📌 暂停工单: %s (ID=%d), 为新工单让路", order.OrderNo, order.ID)

			// 暂停时累加用时
			pauseUpdates := map[string]interface{}{"status": 2}
			if order.CurrentStartTime != nil {
				elapsed := int(time.Since(*order.CurrentStartTime).Seconds())
				pauseUpdates["used_seconds"] = order.UsedSeconds + elapsed
				pauseUpdates["current_start_time"] = nil
			}

			if err := tx.Model(&models.ProOrder{}).
				Where("id = ?", order.ID).
				Updates(pauseUpdates).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("暂停工单失败: %w", err)
			}
		}
	}

	// 2. 结束该设备所有未结束的运行记录
	now := time.Now()
	if err := tx.Model(&models.ProProductionRun{}).
		Where("device_id = ? AND end_time IS NULL", req.DeviceID).
		Update("end_time", now).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("结束旧运行记录失败: %w", err)
	}

	// 3. 验证工单存在且可以开工
	var order models.ProOrder
	if err := tx.First(&order, req.OrderID).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("工单不存在: %w", err)
	}

	// 4. 将操作员ID数组转为JSON字符串
	operatorIDsJSON, err := json.Marshal(req.OperatorIDs)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("操作员ID序列化失败: %w", err)
	}

	// 5. 创建新的运行记录
	run := &models.ProProductionRun{
		OrderID:     req.OrderID,
		DeviceID:    req.DeviceID,
		TeamID:      req.TeamID,
		RunOkQty:    0,
		RunNgQty:    0,
		StartTime:   now,
		EndTime:     nil,
		OperatorIDs: string(operatorIDsJSON),
		Remark:      req.Remark,
	}

	if err := tx.Create(run).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("创建运行记录失败: %w", err)
	}

	// 6. 更新工单状态为"生产中"
	updates := map[string]interface{}{
		"status":             1,   // 生产中
		"current_start_time": now, // 重置当前开始时间
	}
	// 如果是首次开工，记录开工时间
	if order.StartTime == nil {
		updates["start_time"] = now
	}

	if err := tx.Model(&models.ProOrder{}).Where("id = ?", req.OrderID).Updates(updates).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("更新工单状态失败: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	// 重新加载包含关联数据
	DB.Preload("Order").Preload("Team").First(run, run.ID)

	log.Printf("[Database] ✅ 智能开工成功: 工单=%s, 设备=%d, 班组=%d, 运行记录ID=%d",
		order.OrderNo, req.DeviceID, req.TeamID, run.ID)

	return run, nil
}

// ========================================================
// 数据库迁移（自动创建表）
// ========================================================
// 注意: MigrateMESTables 和 InitMESDatabase 已移至 mes_migration.go

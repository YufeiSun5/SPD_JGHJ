package models

import (
	"time"
)

// ========================================================
// 0. 设备表 (复用已有表结构)
// ========================================================

// SysDevice 设备表（映射到已有的 sys_devices 表）
type SysDevice struct {
	ID          int     `gorm:"primaryKey;autoIncrement" json:"id"`
	GatewayID   int     `gorm:"column:gateway_id" json:"gateway_id"`
	DeviceCode  string  `gorm:"column:device_code;type:varchar(50)" json:"device_code"`
	DeviceName  string  `gorm:"column:device_name;type:varchar(100)" json:"device_name"`
	IdentifyKey *string `gorm:"column:identify_key;type:varchar(50)" json:"identify_key"`
}

func (SysDevice) TableName() string {
	return "sys_devices"
}

// ========================================================
// 0.1 设备状态表 (运行状态历史)
// ========================================================

// SysDeviceStatus 设备状态历史记录表
type SysDeviceStatus struct {
	ID          int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	DeviceID    int        `gorm:"not null;index:idx_device_status;index:idx_device_active;comment:设备ID" json:"device_id"`
	Status      int8       `gorm:"not null;default:0;comment:设备状态: 0-空闲, 1-运行, 2-故障" json:"status"`
	StartTime   time.Time  `gorm:"not null;index:idx_device_status;comment:状态开始时间" json:"start_time"`
	EndTime     *time.Time `gorm:"index:idx_device_active;comment:状态结束时间(NULL表示当前状态)" json:"end_time"`
	DurationMin int        `gorm:"default:0;comment:状态持续时长(分钟)" json:"duration_min"`
	ExtraData   *string    `gorm:"type:json;comment:扩展数据(JSON格式,存储温湿度等)" json:"extra_data"`
	Remark      *string    `gorm:"type:varchar(255);comment:备注" json:"remark"`

	// 关联查询用
	Device *SysDevice `gorm:"foreignKey:DeviceID;references:ID" json:"device,omitempty"`
}

func (SysDeviceStatus) TableName() string {
	return "sys_device_status"
}

// DeviceStatusSummary 设备状态统计（用于响应）
type DeviceStatusSummary struct {
	DeviceID      int        `json:"device_id"`
	DeviceName    string     `json:"device_name"`
	CurrentStatus int8       `json:"current_status"` // 当前状态
	StatusName    string     `json:"status_name"`    // 状态名称
	StartTime     *time.Time `json:"start_time"`     // 当前状态开始时间
	DurationMin   int        `json:"duration_min"`   // 当前状态持续时长
	RunningMin    int        `json:"running_min"`    // 今日运行时长
	IdleMin       int        `json:"idle_min"`       // 今日空闲时长
	FaultMin      int        `json:"fault_min"`      // 今日故障时长
	Utilization   float64    `json:"utilization"`    // 利用率(%)
}

// ========================================================
// 1. 班组表 (基础信息)
// ========================================================

// SysTeam 班组表
type SysTeam struct {
	ID         int       `gorm:"primaryKey;autoIncrement" json:"id"`
	TeamName   string    `gorm:"type:varchar(50);not null" json:"team_name"`
	LeaderName *string   `gorm:"type:varchar(50)" json:"leader_name"`
	Status     int8      `gorm:"default:1;comment:状态: 1-启用, 0-禁用" json:"status"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (SysTeam) TableName() string {
	return "sys_teams"
}

// ========================================================
// 2. 人员表 (基础信息 - 仅记录当前状态)
// ========================================================

// SysStaff 员工人员名册
type SysStaff struct {
	ID            int       `gorm:"primaryKey;autoIncrement" json:"id"`
	StaffCode     string    `gorm:"type:varchar(50);not null;unique;comment:工号" json:"staff_code"`
	Name          string    `gorm:"type:varchar(50);not null" json:"name"`
	CurrentTeamID *int      `gorm:"index:idx_staff_team;comment:当前所属班组ID" json:"current_team_id"`
	IsActive      int8      `gorm:"index:idx_staff_active;default:1;comment:在职状态: 1-在职, 0-离职" json:"is_active"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`

	// 关联查询用 (不存入数据库)
	CurrentTeam *SysTeam `gorm:"foreignKey:CurrentTeamID;references:ID" json:"current_team,omitempty"`
}

func (SysStaff) TableName() string {
	return "sys_staff"
}

// ========================================================
// 3. 人员流动历史表 (日志表 - 回溯谁干的活)
// ========================================================

// SysStaffHistory 人员调动历史记录
type SysStaffHistory struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	StaffID      int       `gorm:"not null;index:idx_staff_time" json:"staff_id"`
	TeamID       int       `gorm:"not null;index:idx_team_time" json:"team_id"`
	ActionType   int8      `gorm:"not null;comment:变动类型: 1-加入班组, 2-离开/调出" json:"action_type"`
	HappenedAt   time.Time `gorm:"autoCreateTime;index:idx_staff_time;index:idx_team_time" json:"happened_at"`
	OperatorName *string   `gorm:"type:varchar(50);comment:操作人" json:"operator_name"`

	// 关联查询用
	Staff *SysStaff `gorm:"foreignKey:StaffID;references:ID" json:"staff,omitempty"`
	Team  *SysTeam  `gorm:"foreignKey:TeamID;references:ID" json:"team,omitempty"`
}

func (SysStaffHistory) TableName() string {
	return "sys_staff_history"
}

// ========================================================
// 4. 工单主表 (计划与汇总)
// ========================================================

// ProOrder 生产工单主表
type ProOrder struct {
	ID               int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderNo          string     `gorm:"type:varchar(50);not null;unique;comment:工单号" json:"order_no"`
	ProductCode      string     `gorm:"type:varchar(50);not null;comment:产品型号" json:"product_code"`
	TargetDeviceID   *int       `gorm:"index:idx_order_device;comment:计划生产设备ID" json:"target_device_id"`
	PlanQty          int        `gorm:"not null;default:0;comment:计划生产总数" json:"plan_qty"`
	ActualQty        int        `gorm:"default:0;comment:实际总产出" json:"actual_qty"`
	OkQty            int        `gorm:"default:0;comment:良品总数" json:"ok_qty"`
	NgQty            int        `gorm:"default:0;comment:不良品总数" json:"ng_qty"`
	Status           int8       `gorm:"index:idx_order_status;default:0;comment:状态: 0-待产, 1-生产中, 2-暂停, 3-完工, 4-关闭" json:"status"`
	StartTime        *time.Time `gorm:"comment:首次开工时间" json:"start_time"`
	EndTime          *time.Time `gorm:"comment:最终完工时间" json:"end_time"`
	UsedSeconds      int        `gorm:"default:0;comment:已使用秒数(累计)" json:"used_seconds"`
	CurrentStartTime *time.Time `gorm:"comment:当前开始时间(用于计算本次用时)" json:"current_start_time"`
	Version          int        `gorm:"default:0;comment:乐观锁版本号" json:"version"`
	CreatedAt        time.Time  `gorm:"index:idx_order_created;autoCreateTime" json:"created_at"`
}

func (ProOrder) TableName() string {
	return "pro_orders"
}

// ========================================================
// 5. 生产运行记录表 (执行层 - 分设备、分班次)
// ========================================================

// ProProductionRun 工单分班次/分设备执行记录
type ProProductionRun struct {
	ID          int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID     int64      `gorm:"not null;index:idx_order;comment:关联工单ID" json:"order_id"`
	DeviceID    int        `gorm:"not null;index:idx_run_device_time;comment:执行设备ID" json:"device_id"`
	TeamID      int        `gorm:"not null;index:idx_run_team;comment:执行班组ID" json:"team_id"`
	RunOkQty    int        `gorm:"default:0;comment:本班次增量-良品" json:"run_ok_qty"`
	RunNgQty    int        `gorm:"default:0;comment:本班次增量-不良品" json:"run_ng_qty"`
	StartTime   time.Time  `gorm:"not null;index:idx_run_device_time" json:"start_time"`
	EndTime     *time.Time `gorm:"index:idx_run_active;comment:本班次结束时间" json:"end_time"`
	OperatorIDs string     `gorm:"type:json;comment:人员快照" json:"operator_ids"` // JSON数组字符串
	Remark      *string    `gorm:"type:varchar(255)" json:"remark"`

	// 关联查询用
	Order *ProOrder `gorm:"foreignKey:OrderID;references:ID" json:"order,omitempty"`
	Team  *SysTeam  `gorm:"foreignKey:TeamID;references:ID" json:"team,omitempty"`
}

func (ProProductionRun) TableName() string {
	return "pro_production_runs"
}

// ========================================================
// DTO (数据传输对象 - 用于API请求/响应)
// ========================================================

// CreateTeamRequest 创建班组请求
type CreateTeamRequest struct {
	TeamName   string  `json:"team_name" binding:"required"`
	LeaderName *string `json:"leader_name"`
}

// UpdateTeamRequest 更新班组请求
type UpdateTeamRequest struct {
	TeamName   *string `json:"team_name"`
	LeaderName *string `json:"leader_name"`
	Status     *int8   `json:"status"`
}

// CreateStaffRequest 创建员工请求
type CreateStaffRequest struct {
	StaffCode     string `json:"staff_code" binding:"required"`
	Name          string `json:"name" binding:"required"`
	CurrentTeamID *int   `json:"current_team_id"`
	IsActive      *int8  `json:"is_active"`
}

// UpdateStaffRequest 更新员工请求
type UpdateStaffRequest struct {
	Name          *string `json:"name"`
	CurrentTeamID *int    `json:"current_team_id"`
	IsActive      *int8   `json:"is_active"`
}

// TransferStaffRequest 调动员工请求
type TransferStaffRequest struct {
	StaffID      int     `json:"staff_id" binding:"required"`
	NewTeamID    int     `json:"new_team_id" binding:"required"`
	OperatorName *string `json:"operator_name"`
}

// CreateOrderRequest 创建工单请求
type CreateOrderRequest struct {
	OrderNo        string `json:"order_no" binding:"required"`
	ProductCode    string `json:"product_code" binding:"required"`
	TargetDeviceID *int   `json:"target_device_id"`
	PlanQty        int    `json:"plan_qty" binding:"required"`
}

// UpdateOrderRequest 更新工单请求
type UpdateOrderRequest struct {
	ProductCode    *string `json:"product_code"`
	TargetDeviceID *int    `json:"target_device_id"`
	PlanQty        *int    `json:"plan_qty"`
	Status         *int8   `json:"status"`
}

// StartProductionRequest 开始生产请求
type StartProductionRequest struct {
	OrderID     int64   `json:"order_id" binding:"required"`
	DeviceID    int     `json:"device_id" binding:"required"`
	TeamID      int     `json:"team_id" binding:"required"`
	OperatorIDs []int   `json:"operator_ids" binding:"required"`
	Remark      *string `json:"remark"`
}

// UpdateProductionRunRequest 更新生产运行记录
type UpdateProductionRunRequest struct {
	RunOkQty *int `json:"run_ok_qty"`
	RunNgQty *int `json:"run_ng_qty"`
}

// EndProductionRequest 结束生产请求
type EndProductionRequest struct {
	RunID  int64  `json:"run_id" binding:"required"`
	Remark string `json:"remark"`
}

// OrderSummaryResponse 工单汇总响应
type OrderSummaryResponse struct {
	ProOrder
	RunCount      int                `json:"run_count"`      // 总运行次数
	TotalDuration float64            `json:"total_duration"` // 总运行时长(小时)
	AvgEfficiency float64            `json:"avg_efficiency"` // 平均效率(实际/计划)
	Runs          []ProProductionRun `json:"runs,omitempty"` // 运行记录明细
}

// ========================================================
// 6. 设备登录/班次记录表 (考勤表 - 独立于工单)
// ========================================================

// ProMachineSession 设备登录与班次记录表
type ProMachineSession struct {
	ID          int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	DeviceID    int        `gorm:"not null;index:idx_session_device_logout;comment:登录设备ID" json:"device_id"`
	TeamID      int        `gorm:"not null;index:idx_session_team_time;comment:登录班组ID" json:"team_id"`
	StaffIDs    string     `gorm:"type:json;not null;comment:当班员工ID列表" json:"staff_ids"` // JSON数组: [101,102,105]
	LoginTime   time.Time  `gorm:"not null;index:idx_session_team_time;index:idx_session_login" json:"login_time"`
	LogoutTime  *time.Time `gorm:"index:idx_session_device_logout;comment:下班/登出时间(NULL代表正在上班)" json:"logout_time"`
	DurationMin int        `gorm:"default:0;comment:上班时长(分钟)" json:"duration_min"`

	// 关联查询用
	Team *SysTeam `gorm:"foreignKey:TeamID;references:ID" json:"team,omitempty"`
}

func (ProMachineSession) TableName() string {
	return "pro_machine_sessions"
}

// ========================================================
// DTO (数据传输对象 - 设备登录相关)
// ========================================================

// LoginRequest 设备登录请求
type LoginRequest struct {
	DeviceID int   `json:"device_id" binding:"required"`
	TeamID   int   `json:"team_id" binding:"required"`
	StaffIDs []int `json:"staff_ids" binding:"required"` // 当班员工ID列表
}

// LogoutRequest 设备登出请求
type LogoutRequest struct {
	DeviceID int     `json:"device_id" binding:"required"`
	Remark   *string `json:"remark"` // 下班备注
}

// SessionStatusResponse 设备登录状态响应
type SessionStatusResponse struct {
	ProMachineSession
	IsActive   bool        `json:"is_active"`  // 是否正在上班
	StaffList  []*SysStaff `json:"staff_list"` // 当班员工详情
	WorkedMin  int         `json:"worked_min"` // 已工作分钟数
	IdleMin    int         `json:"idle_min"`   // 空闲分钟数
	Efficiency float64     `json:"efficiency"` // 工时利用率
}

// ========================================================
// DTO (数据传输对象 - 设备状态相关)
// ========================================================

// UpdateDeviceStatusRequest 更新设备状态请求
type UpdateDeviceStatusRequest struct {
	DeviceID int     `json:"device_id" binding:"required"`
	Status   int8    `json:"status" binding:"required,min=0,max=2"` // 0-空闲, 1-运行, 2-故障
	Remark   *string `json:"remark"`
}

// DeviceStatusQueryRequest 设备状态查询请求
type DeviceStatusQueryRequest struct {
	DeviceID  *int       `form:"device_id"`
	Status    *int8      `form:"status"`
	StartTime *time.Time `form:"start_time"`
	EndTime   *time.Time `form:"end_time"`
}

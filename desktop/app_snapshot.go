package main

// ========================================================
// 班次生产快照：自动生成 + 查询 + 手动重建（Wails 绑定）
// Shift Production Snapshot: auto-generation, query, and manual rebuild (Wails bindings)
// シフト生産スナップショット：自動生成・クエリ・手動再生成（Wailsバインド）
//
// 数据流：
//   定时器（StartSnapshotTicker）→ checkAndGenerateSnapshots →
//     对每个 schedule × shift × device 调用 generateOneSnapshot →
//       聚合 sys_device_status / sys_data_history / pro_machine_sessions / sys_staff_history
//       → INSERT pro_shift_snapshots
//
//   前端查询 → GetShiftSnapshots（按日期/时间安排/设备/班组/人员筛选）
//   手动重建 → RegenerateShiftSnapshot（删旧记录后重新聚合）
// ========================================================

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"gin-mqtt-pgsql/database"
	"gin-mqtt-pgsql/models"
)

// ─── 快照定时器 ───────────────────────────────────────────

var (
	snapshotStopCh chan struct{}
	snapshotOnce   sync.Once
	// snapshotRunning 用原子标志防止同一时刻两次并发调用 checkAndGenerateSnapshots
	// Atomic flag to prevent concurrent calls to checkAndGenerateSnapshots
	snapshotRunning int32
)

// StartSnapshotTicker 启动班次快照定时检查器（每 60 秒一次）。
// 在 main.go initBackend 尾部调用一次即可；内部保证幂等。
// Starts the shift snapshot ticker (every 60s). Called once from initBackend.
// シフトスナップショットの定期チェッカーを起動（60秒ごと）。initBackendから1回呼ぶ。
func StartSnapshotTicker(app *App) {
	snapshotOnce.Do(func() {
		snapshotStopCh = make(chan struct{})
		go func() {
			ticker := time.NewTicker(60 * time.Second)
			defer ticker.Stop()
			// 启动后立即检查一次（追补遗漏）
			checkAndGenerateSnapshots(app)
			for {
				select {
				case <-ticker.C:
					checkAndGenerateSnapshots(app)
				case <-snapshotStopCh:
					log.Println("[SnapshotTicker] stopped")
					return
				}
			}
		}()
		log.Println("[SnapshotTicker] started (interval=60s)")
	})
}

// checkAndGenerateSnapshots 扫描所有 schedule → shift，生成已结束但尚未快照的记录。
// 同时追补：若逻辑日为昨天，且昨天的跨零点班次也没快照，一并补全。
//
// CN: 使用原子标志避免重入（定时器与启动时的立即调用并发时不会重复生成）。
// EN: Uses atomic flag to prevent re-entrancy (timer vs. startup call won't overlap).
// JP: 原子フラグで再入を防止（タイマーと起動時の即時呼び出しが重複しないようにする）。
func checkAndGenerateSnapshots(app *App) {
	if database.DB == nil {
		return
	}
	// 若已有一次在执行则跳过本次，避免并发重复写入
	if !atomic.CompareAndSwapInt32(&snapshotRunning, 0, 1) {
		return
	}
	defer atomic.StoreInt32(&snapshotRunning, 0)
	schedules, err := app.GetShiftSchedules()
	if err != nil {
		log.Printf("[Snapshot] 获取时间安排组失败: %v", err)
		return
	}

	now := time.Now()

	// CN: 全局 CT 只作为设备未配置独立 cycle_time 时的兜底值。
	// EN: Global CT is only the fallback when a device has no cycle_time override.
	// JP: グローバル CT は設備別 cycle_time が未設定の場合のみフォールバックとして使う。
	globalCT := app.getGlobalCycleTimeFallback()

	// CN: 快照按设备读取 schedule_id 和 cycle_time，确保两台设备可使用不同 CT。
	// EN: Snapshots read schedule_id and cycle_time per device so two devices can use different CT values.
	// JP: スナップショットは設備ごとに schedule_id と cycle_time を読み、2台設備で異なる CT を使えるようにする。
	var allDevices []models.SysDevice
	database.DB.Find(&allDevices)
	schedDevMap := map[int][]models.SysDevice{}
	for _, d := range allDevices {
		if d.ScheduleID != nil {
			schedDevMap[*d.ScheduleID] = append(schedDevMap[*d.ScheduleID], d)
		}
	}

	// CN: 需要检查的逻辑日期：前天、昨天、今天（3天，确保跨零点班次不漏检）。
	// EN: Logical days to check: day-2, day-1, today (3 days catches cross-midnight shifts).
	// JP: チェックする論理日：一昨日・昨日・今日（3日間で深夜またぎシフトの漏れを防ぐ）。
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	dates := []time.Time{today.AddDate(0, 0, -2), today.AddDate(0, 0, -1), today}

	for _, sched := range schedules {
		if !sched.IsActive {
			continue
		}
		devices := schedDevMap[sched.ID]
		if len(devices) == 0 {
			continue
		}

		// CN: 求逻辑日临界时刻——找 sort_order 最小的活跃班次，其开始时间（分钟）即为临界。
		//   开始时间早于临界的班次（如三班 0:00 < 一班 8:00）实际运行于日历次日，
		//   但属于本逻辑日；生成快照时 snapshot_date 应用逻辑日，shift_start/end 用日历次日。
		// EN: Logical day boundary = start time (minutes) of active shift with min sort_order.
		//   Shifts starting before this boundary (e.g. 三班 0:00 < 一班 8:00) run on the NEXT
		//   calendar day but belong to THIS logical day. snapshot_date uses the logical day date.
		// JP: 論理日境界 = sort_order 最小の活動シフトの開始時刻（分）。境界より前に始まるシフト
		//   （三班 0:00 < 一班 8:00）は翌カレンダー日に動くが、この論理日に属する。
		logicalBoundaryMin := -1
		minSortOrder := -1
		for _, s := range sched.Shifts {
			if !s.IsActive {
				continue
			}
			if minSortOrder < 0 || s.SortOrder < minSortOrder {
				minSortOrder = s.SortOrder
				logicalBoundaryMin = s.StartHour*60 + s.StartMin
			}
		}
		if logicalBoundaryMin < 0 {
			continue // 无活跃班次
		}

		for _, shift := range sched.Shifts {
			if !shift.IsActive {
				continue
			}

			// CN: 若班次开始时刻（分钟）早于逻辑日临界，则该班次在日历次日运行（offset=1）。
			//   例：三班 start=0:00(0) < 临界 8:00(480) → calendarDayOffset=1。
			// EN: If shift start < logical boundary, shift runs on logical_day+1 (offset=1).
			// JP: シフト開始が論理日境界より前なら、論理日の翌カレンダー日に動く（offset=1）。
			shiftStartMin := shift.StartHour*60 + shift.StartMin
			calendarDayOffset := 0
			if shiftStartMin < logicalBoundaryMin {
				calendarDayOffset = 1
			}

			for _, logicalDate := range dates {
				// 实际时间戳基于 calendarBase；snapshot_date 始终用逻辑日
				// Timestamps from calendarBase; snapshot_date always uses logical day.
				calendarBase := logicalDate.AddDate(0, 0, calendarDayOffset)
				shiftStart, shiftEnd := resolveShiftDatetime(shift, calendarBase)
				if shiftEnd.After(now) {
					continue // 班次尚未结束
				}
				dateStr := logicalDate.Format("2006-01-02")

				for _, dev := range devices {
					if snapshotExists(dateStr, dev.ID, shift.ID) {
						continue
					}
					ct := globalCT
					if dev.CycleTime != nil && *dev.CycleTime > 0 {
						ct = *dev.CycleTime
					}
					if err := generateOneSnapshot(dev, shift, sched.ID, dateStr, shiftStart, shiftEnd, ct); err != nil {
						log.Printf("[Snapshot] 生成快照失败 date=%s dev=%d shift=%d: %v", dateStr, dev.ID, shift.ID, err)
					}
				}
			}
		}
	}
}

// resolveShiftDatetime 将 ShiftConfig + 基准日期 → 实际的 start/end time.Time
func resolveShiftDatetime(s ShiftConfig, base time.Time) (time.Time, time.Time) {
	start := base.Add(time.Duration(s.StartHour*60+s.StartMin) * time.Minute)
	endMin := s.EndHour*60 + s.EndMin
	startMin := s.StartHour*60 + s.StartMin
	if endMin <= startMin {
		endMin += 24 * 60 // 跨零点
	}
	end := base.Add(time.Duration(endMin) * time.Minute)
	return start, end
}

func snapshotExists(date string, deviceID, shiftID int) bool {
	var count int64
	database.DB.Model(&models.ProShiftSnapshot{}).
		Where("snapshot_date = ? AND device_id = ? AND shift_id = ?", date, deviceID, shiftID).
		Count(&count)
	return count > 0
}

// ─── 单条快照生成核心 ────────────────────────────────────

// StaffSnapshotItem 人员快照条目（含工时拆分，兼容旧格式）
// CN: 改造后新增 TeamID/TeamName/WorkSec/Pct，旧快照这几个字段为零值；前端用 work_sec > 0 区分新旧。
// EN: New fields TeamID/TeamName/WorkSec/Pct added; legacy snapshots will have zero values. Frontend distinguishes by work_sec > 0.
// JP: 新フィールド TeamID/TeamName/WorkSec/Pct を追加。旧スナップショットはゼロ値。フロントは work_sec > 0 で判定。
type StaffSnapshotItem struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Code     string  `json:"code"`
	TeamID   int     `json:"team_id,omitempty"`   // 所属 session 的班组 ID / Team ID of the session / セッションの班組 ID
	TeamName string  `json:"team_name,omitempty"` // 班组名称 / Team name / 班組名称
	WorkSec  int     `json:"work_sec,omitempty"`  // 该员工在本班次的有效工作秒数 / Effective working seconds in this shift / このシフトでの有効勤務秒数
	Pct      float64 `json:"pct,omitempty"`       // 工时占比(%) / Work-time percentage / 工時占比(%)
}

// SessionSnapshotItem 单条 session 在该班次时间窗口内的工时快照
// CN: 记录与班次窗口的有效重叠秒数和占比；多条记录之和 = 总工时（含交叉上下机场景）。
// EN: Records effective overlap seconds and percentage with the shift window; sum of all = total work time.
// JP: シフト時間帯との有効重複秒数と占比を記録。全 session の合計 = 総工時（交替交機シナリオ含む）。
type SessionSnapshotItem struct {
	SessionID  int64   `json:"session_id"`  // pro_machine_sessions.id
	TeamID     int     `json:"team_id"`     // 该 session 的班组 ID / Team ID / 班組 ID
	TeamName   string  `json:"team_name"`   // 班组名称 / Team name / 班組名称
	StaffIDs   []int   `json:"staff_ids"`   // 该 session 的当班人员 ID 列表 / On-duty staff IDs / 当班人員 ID リスト
	LoginTime  string  `json:"login_time"`  // 原始登录时间 "HH:MM" / Original login time / ログイン時刻
	LogoutTime string  `json:"logout_time"` // 原始登出时间 "HH:MM"（进行中为空）/ Logout time, empty if still active / ログアウト時刻（進行中は空）
	OverlapSec int     `json:"overlap_sec"` // 与班次时间窗口的有效重叠秒数 / Effective overlap seconds / 有効重複秒数
	Pct        float64 `json:"pct"`         // 占班次总工时的百分比 / Percentage of total shift work time / シフト総工時に占める割合
}

// BreakSnapshotItem 休息时间快照条目
type BreakSnapshotItem struct {
	Name  string `json:"name"`
	Start string `json:"start"` // "HH:MM"
	End   string `json:"end"`   // "HH:MM"
}

func generateOneSnapshot(dev models.SysDevice, shift ShiftConfig, schedID int, dateStr string, shiftStart, shiftEnd time.Time, ct float64) error {
	snap := models.ProShiftSnapshot{
		SnapshotDate: dateStr,
		DeviceID:     dev.ID,
		DeviceName:   dev.DeviceName,
		ScheduleID:   schedID,
		ShiftID:      shift.ID,
		ShiftName:    shift.Name,
		ShiftStart:   shiftStart,
		ShiftEnd:     shiftEnd,
		CycleTime:    ct,
		// CN: MySQL JSON 列不接受空字符串，无数据时默认写入合法的空数组 "[]"。
		// EN: MySQL JSON columns reject empty strings; default to valid empty JSON array.
		// JP: MySQL の JSON 列は空文字列を拒否するため、データなしの場合は "[]" をデフォルト値とする。
		StaffSnapshot:    "[]",
		SessionsSnapshot: "[]",
	}

	// 1. 休息时间快照 + 理论工作秒数
	// CN: 休息段配置可能存在重叠（同一时间多个休息条目），必须先合并区间再计算净休息分钟数，
	//     否则 plan_work_sec 会被多扣，导致时间稼动率（avail_pct）虚高超过 100%。
	// EN: Break segments may overlap; merge intervals before summing to avoid over-deducting
	//     plan_work_sec, which would artificially inflate avail_pct beyond 100%.
	// JP: 休憩設定が重複する場合があるため、先に区間をマージしてから純休憩分数を計算する。
	//     マージしないと plan_work_sec が過小になり avail_pct が 100% を超えてしまう。
	breakItems := make([]BreakSnapshotItem, len(shift.Breaks))
	breakIvs := make([][2]int, 0, len(shift.Breaks))
	for i, b := range shift.Breaks {
		breakItems[i] = BreakSnapshotItem{
			Name:  b.Name,
			Start: fmt.Sprintf("%02d:%02d", b.StartHour, b.StartMin),
			End:   fmt.Sprintf("%02d:%02d", b.EndHour, b.EndMin),
		}
		bStart := b.StartHour*60 + b.StartMin
		bEnd := b.EndHour*60 + b.EndMin
		if bEnd <= bStart {
			bEnd += 24 * 60
		}
		if bEnd > bStart {
			breakIvs = append(breakIvs, [2]int{bStart, bEnd})
		}
	}
	if bs, err := json.Marshal(breakItems); err == nil {
		snap.BreakConfig = string(bs)
	}

	totalBreakMin := mergeAndSumIntIntervals(breakIvs)
	shiftDurationMin := int(shiftEnd.Sub(shiftStart).Minutes())
	snap.PlanWorkSec = (shiftDurationMin - totalBreakMin) * 60
	if snap.PlanWorkSec < 0 {
		snap.PlanWorkSec = 0
	}

	// 2. 设备状态统计（run/idle/fault 秒数）
	// CN: 按状态类型分组 → 各组内合并重叠区间 → 再求秒数，
	//     防止两个程序实例同时运行写入重叠记录时导致空闲/运行时间虚高。
	// EN: Group by status type, merge overlapping intervals within each group, then sum.
	//     Prevents inflated idle/run time when two app instances wrote duplicate records.
	// JP: ステータス種別でグループ化→重複区間をマージ→合計秒数を計算。
	//     複数インスタンスによる重複レコードで稼動時間が水増しされるのを防ぐ。
	var statusRows []models.SysDeviceStatus
	database.DB.Where("device_id = ? AND start_time < ? AND (end_time IS NULL OR end_time > ?)",
		dev.ID, shiftEnd, shiftStart).Find(&statusRows)

	byStatus := map[int8][][2]time.Time{}
	for _, r := range statusRows {
		ss := r.StartTime
		if ss.Before(shiftStart) {
			ss = shiftStart
		}
		se := shiftEnd
		if r.EndTime != nil && r.EndTime.Before(shiftEnd) {
			se = *r.EndTime
		}
		if se.After(ss) {
			byStatus[r.Status] = append(byStatus[r.Status], [2]time.Time{ss, se})
		}
	}
	snap.DeviceRunSec = sumMergedIntervals(byStatus[1])
	snap.DeviceIdleSec = sumMergedIntervals(byStatus[0])
	snap.DeviceFaultSec = sumMergedIntervals(byStatus[2])

	// 3. 产量统计：调用共享的防复位回跳聚合函数，口径与精确产量统计完全一致。
	// CN: 不再内嵌仅使用 LAG(prev_val) 的弱口径 SQL；统一走 database.GetShiftWindowProduction，
	//     该函数实现了与 GetHourlyProductionAccurate / GetMonthlyProductionAccurate 相同的三档
	//     last_nonzero_val 规则，防止设备复位导致的回跳重计。
	// EN: Replaced the weak inline SQL (LAG-only) with GetShiftWindowProduction, which shares the
	//     same three-branch last_nonzero_val anti-bounce rule used in accurate hourly/monthly stats.
	// JP: prev_val のみの弱口径インライン SQL を廃止し、時間別・月次精確統計と同一の
	//     last_nonzero_val 三分岐規則を持つ GetShiftWindowProduction を使用する。
	oeeConf := hardcodedOEEConfig(dev.ID, ct)
	if oeeConf != nil {
		cfg := database.DeviceVarConfig{
			ProductionVarID: oeeConf.VarOK,
			NgAddVarID:      oeeConf.VarNGAdd,
			NgSubVarID:      oeeConf.VarNGSub,
		}
		if pr, err := database.GetShiftWindowProduction(cfg, shiftStart, shiftEnd); err == nil {
			snap.TotalQty = pr.TotalQty
			snap.OkQty = pr.OkQty
			snap.NgQty = pr.NgQty
		} else {
			log.Printf("[generateOneSnapshot] 产量聚合失败 device=%d shift=%s: %v", dev.ID, dateStr, err)
		}
	}

	// 4. 班组和人员：取该班次时间窗口内所有关联 session，计算各 session 的有效重叠工时
	// CN: 改造前只取一条（First），现在用 Find 取全部，解决多段上/下机时的归属丢失问题。
	// EN: Previously only one session was fetched (First); now Find retrieves all, fixing multi-session attribution loss.
	// JP: 以前は First で1件のみ取得していたが、Find で全件取得に変更し、複数セッションの帰属漏れを修正。
	var sessions []models.ProMachineSession
	database.DB.Preload("Team").
		Where("device_id = ? AND login_time < ? AND (logout_time IS NULL OR logout_time > ?)",
			dev.ID, shiftEnd, shiftStart).
		Order("login_time ASC").
		Find(&sessions)

	if len(sessions) > 0 {
		// ── 计算每条 session 与班次窗口的有效重叠秒数 ──────────────
		type sessionOverlap struct {
			Session    models.ProMachineSession
			OverlapSec int
		}
		var overlaps []sessionOverlap
		totalOverlapSec := 0

		for _, sess := range sessions {
			effStart := sess.LoginTime
			if effStart.Before(shiftStart) {
				effStart = shiftStart
			}
			effEnd := shiftEnd
			if sess.LogoutTime != nil && sess.LogoutTime.Before(shiftEnd) {
				effEnd = *sess.LogoutTime
			}
			sec := int(effEnd.Sub(effStart).Seconds())
			if sec <= 0 {
				continue
			}
			overlaps = append(overlaps, sessionOverlap{Session: sess, OverlapSec: sec})
			totalOverlapSec += sec
		}

		if len(overlaps) > 0 {
			// ── 构建 sessions_snapshot JSON ──────────────────────────
			sessionItems := make([]SessionSnapshotItem, len(overlaps))
			for i, o := range overlaps {
				pct := 0.0
				if totalOverlapSec > 0 {
					pct = math.Round(float64(o.OverlapSec)*10000/float64(totalOverlapSec)) / 100
				}
				teamName := ""
				if o.Session.Team != nil {
					teamName = o.Session.Team.TeamName
				}
				logoutStr := ""
				if o.Session.LogoutTime != nil {
					logoutStr = o.Session.LogoutTime.Format("15:04")
				}
				var sIDs []int
				json.Unmarshal([]byte(o.Session.StaffIDs), &sIDs)
				sessionItems[i] = SessionSnapshotItem{
					SessionID:  o.Session.ID,
					TeamID:     o.Session.TeamID,
					TeamName:   teamName,
					StaffIDs:   sIDs,
					LoginTime:  o.Session.LoginTime.Format("15:04"),
					LogoutTime: logoutStr,
					OverlapSec: o.OverlapSec,
					Pct:        pct,
				}
			}
			if bs, err := json.Marshal(sessionItems); err == nil {
				snap.SessionsSnapshot = string(bs)
			}

			// ── 构建 staff_snapshot JSON（含工时）───────────────────
			// CN: 同一人出现在多个 session 中时，累加工时（极罕见，但正确处理）
			// EN: If the same staff appears in multiple sessions, accumulate their work time.
			// JP: 同一人員が複数セッションに出現する場合は工時を累積する。
			staffMap := map[int]*StaffSnapshotItem{}
			for _, o := range overlaps {
				var sIDs []int
				json.Unmarshal([]byte(o.Session.StaffIDs), &sIDs)
				teamName := ""
				if o.Session.Team != nil {
					teamName = o.Session.Team.TeamName
				}
				for _, sid := range sIDs {
					if existing, ok := staffMap[sid]; ok {
						existing.WorkSec += o.OverlapSec
					} else {
						staffMap[sid] = &StaffSnapshotItem{
							ID:       sid,
							TeamID:   o.Session.TeamID,
							TeamName: teamName,
							WorkSec:  o.OverlapSec,
						}
					}
				}
			}
			// 批量查人员姓名和工号
			if len(staffMap) > 0 {
				allStaffIDs := make([]int, 0, len(staffMap))
				for sid := range staffMap {
					allStaffIDs = append(allStaffIDs, sid)
				}
				var staffList []models.SysStaff
				database.DB.Where("id IN ?", allStaffIDs).Find(&staffList)
				for _, s := range staffList {
					if item, ok := staffMap[s.ID]; ok {
						item.Name = s.Name
						item.Code = s.StaffCode
					}
				}
				// 计算占比并序列化
				staffItems := make([]StaffSnapshotItem, 0, len(staffMap))
				for _, item := range staffMap {
					if totalOverlapSec > 0 {
						item.Pct = math.Round(float64(item.WorkSec)*10000/float64(totalOverlapSec)) / 100
					}
					staffItems = append(staffItems, *item)
				}
				sort.Slice(staffItems, func(i, j int) bool {
					return staffItems[i].WorkSec > staffItems[j].WorkSec
				})
				if bs, err := json.Marshal(staffItems); err == nil {
					snap.StaffSnapshot = string(bs)
				}
			}

			// ── 主班组取工时最长的 session（兼容旧字段 team_id/team_name/session_id）──
			// CN: 保留这三个字段供 GetShiftSnapshots WHERE team_id = ? 筛选使用。
			// EN: Keep these fields for the WHERE team_id = ? filter in GetShiftSnapshots.
			// JP: GetShiftSnapshots の WHERE team_id = ? フィルタのためこれらのフィールドを保持。
			maxOverlap := overlaps[0]
			for _, o := range overlaps[1:] {
				if o.OverlapSec > maxOverlap.OverlapSec {
					maxOverlap = o
				}
			}
			snap.TeamID = &maxOverlap.Session.TeamID
			if maxOverlap.Session.Team != nil {
				snap.TeamName = maxOverlap.Session.Team.TeamName
			}
			snap.SessionID = &maxOverlap.Session.ID
		}
	}

	// 5. 计算 OEE 四项指标
	computeOEE(&snap)

	// CN: 数据库唯一索引兜底：若并发/重启导致极罕见的重复写入，直接忽略重复键错误。
	// EN: DB unique index as final guard: silently ignore duplicate-key error (rare race condition).
	// JP: DB ユニークインデックスで最終防御。重複キーエラーは無視（極めてまれな競合状態）。
	if err := database.DB.Create(&snap).Error; err != nil {
		if strings.Contains(err.Error(), "1062") || strings.Contains(err.Error(), "Duplicate entry") ||
			strings.Contains(err.Error(), "UNIQUE constraint") {
			log.Printf("[Snapshot] 跳过重复快照 date=%s dev=%d shift=%d", snap.SnapshotDate, snap.DeviceID, snap.ShiftID)
			return nil
		}
		return err
	}
	return nil
}

// mergeAndSumIntIntervals 合并重叠整数区间并返回累加长度（单位与输入一致，通常为分钟）。
//
// CN: 用于休息段配置去重：将多个 [startMin, endMin) 区间排序后合并，再求总分钟数。
//
//	避免重叠休息段被重复计入 plan_work_sec 导致理论工作时间偏小。
//
// EN: Merge overlapping [start, end) int intervals and return total length.
//
//	Used to deduplicate overlapping break segments when computing plan_work_sec.
//
// JP: 重複する [start, end) 整数区間をマージし合計長を返す。休憩設定の重複除去に使用する。
func mergeAndSumIntIntervals(ivs [][2]int) int {
	if len(ivs) == 0 {
		return 0
	}
	sort.Slice(ivs, func(i, j int) bool { return ivs[i][0] < ivs[j][0] })
	merged := [][2]int{ivs[0]}
	for _, x := range ivs[1:] {
		last := &merged[len(merged)-1]
		if x[0] <= last[1] { // 重叠或相邻
			if x[1] > last[1] {
				last[1] = x[1]
			}
		} else {
			merged = append(merged, x)
		}
	}
	total := 0
	for _, m := range merged {
		total += m[1] - m[0]
	}
	return total
}

// sumMergedIntervals 对一组时间区间去重合并后求总秒数。
//
// CN: 先按起始时间排序，再将重叠/相邻区间合并，最后求各合并段的秒数之和。
//
//	目的是解决「两个程序实例同时运行时，同一时段被写入多条状态记录」导致累加重复的问题。
//
// EN: Sort by start, merge overlapping intervals, then sum durations.
//
//	Prevents double-counting when two app instances wrote overlapping status records.
//
// JP: 開始時刻でソートし、重複区間をマージしてから合計秒数を求める。
//
//	複数インスタンスが同一時間帯に重複レコードを書いた場合の二重計上を防ぐ。
func sumMergedIntervals(ivs [][2]time.Time) int {
	if len(ivs) == 0 {
		return 0
	}
	sort.Slice(ivs, func(i, j int) bool { return ivs[i][0].Before(ivs[j][0]) })
	merged := [][2]time.Time{ivs[0]}
	for _, x := range ivs[1:] {
		last := &merged[len(merged)-1]
		if !x[0].After(last[1]) { // 重叠或相邻
			if x[1].After(last[1]) {
				last[1] = x[1]
			}
		} else {
			merged = append(merged, x)
		}
	}
	total := 0
	for _, m := range merged {
		total += int(m[1].Sub(m[0]).Seconds())
	}
	return total
}

func computeOEE(s *models.ProShiftSnapshot) {
	// 时间稼动率 = 实际运行时间 / 理论工作时间
	if s.PlanWorkSec > 0 {
		s.AvailPct = math.Round(float64(s.DeviceRunSec)*10000/float64(s.PlanWorkSec)) / 100
	}
	// 性能稼动率 = (总产量 × CT) / 实际运行时间
	if s.DeviceRunSec > 0 && s.CycleTime > 0 {
		s.PerfPct = math.Round(float64(s.TotalQty)*s.CycleTime*10000/float64(s.DeviceRunSec)) / 100
	}
	// 直通率 = 良品 / 总产量
	if s.TotalQty > 0 {
		s.QualityPct = math.Round(float64(s.OkQty)*10000/float64(s.TotalQty)) / 100
	} else {
		s.QualityPct = 100
	}
	// OEE = 时间稼动率 × 性能稼动率 × 直通率 / 10000
	// CN: 三项指标均可超过 100%：
	//   - 时间稼动率 > 100% → 设备实际运行时间超过理论工时（如提前开班、延迟停班）
	//   - 性能稼动率 > 100% → 实际节拍优于理论节拍（设备效率高于基准）
	//   - OEE > 100%        → 上述两者共同作用的结果，属于正常现象，不截断
	// EN: All three metrics can legally exceed 100% and are stored as-is.
	// JP: 3指標はいずれも 100% を超えることがあり、そのまま保持する。
	s.OeePct = math.Round(s.AvailPct*s.PerfPct*s.QualityPct) / 10000
}

// ─── Wails 查询接口 ──────────────────────────────────────

// GetShiftSnapshots 按日期范围+可选时间安排/设备/班组/人员筛选班次生产快照
// Query shift production snapshots by date range with optional schedule/device/team/staff filters.
// 日付範囲・スケジュール・設備・班組・人員でシフト生産スナップショットを検索する。
// GetShiftSnapshots 查询班次生产快照列表
//
// CN: startDate/endDate 支持两种格式：
//   - "YYYY-MM-DD"（纯日期）→ 用 snapshot_date 过滤
//   - "YYYY-MM-DDTHH:MM"（含时刻，前端 datetime-local 格式）→ 用 shift_start/shift_end 做分钟级过滤
//     scheduleID/deviceID/teamID 为 nil 或 <=0 时不过滤；staffName 非空时在 staff_snapshot 中模糊匹配。
//
// EN: startDate/endDate accept "YYYY-MM-DD" (date-only → filter by snapshot_date) or
//
//	"YYYY-MM-DDTHH:MM" (datetime → filter by shift_start/shift_end for minute-level precision).
//	Nil/non-positive scheduleID/deviceID/teamID means no filter; non-empty staffName uses LIKE on staff_snapshot.
//
// JP: startDate/endDate は "YYYY-MM-DD"（日付のみ→snapshot_dateでフィルタ）または
//
//	"YYYY-MM-DDTHH:MM"（日時→shift_start/shift_end で分単位フィルタ）を受け付ける。
//	scheduleID/deviceID/teamID が nil または 0 以下なら未指定扱い。staffName は staff_snapshot を LIKE 検索する。
func (a *App) GetShiftSnapshots(startDate, endDate string, scheduleID *int, deviceID *int, teamID *int, staffName string) ([]models.ProShiftSnapshot, error) {
	if database.DB == nil {
		return nil, fmt.Errorf("数据库未连接")
	}
	// CN: 同一日期内按 sys_shifts.sort_order DESC 排列班次（夜班最前），再按设备排
	// EN: Within the same date, sort shifts by sys_shifts.sort_order DESC (latest shift first), then by device.
	// JP: 同一日付内は sys_shifts.sort_order DESC でシフトを並べ（夜班が先頭）、次に設備でソート。
	query := database.DB.Order("snapshot_date DESC, (SELECT sort_order FROM sys_shifts WHERE id = shift_id) DESC, device_id ASC")

	// 含 "T" 表示前端传来的是 datetime-local 格式（分钟级），用 shift_start / shift_end 过滤
	// No "T" → pure date → filter on snapshot_date
	if startDate != "" {
		if strings.Contains(startDate, "T") {
			query = query.Where("shift_start >= ?", strings.ReplaceAll(startDate, "T", " "))
		} else {
			query = query.Where("snapshot_date >= ?", startDate)
		}
	}
	if endDate != "" {
		if strings.Contains(endDate, "T") {
			query = query.Where("shift_end <= ?", strings.ReplaceAll(endDate, "T", " "))
		} else {
			query = query.Where("snapshot_date <= ?", endDate)
		}
	}
	if scheduleID != nil && *scheduleID > 0 {
		query = query.Where("schedule_id = ?", *scheduleID)
	}
	if deviceID != nil && *deviceID > 0 {
		query = query.Where("device_id = ?", *deviceID)
	}
	if teamID != nil && *teamID > 0 {
		query = query.Where("team_id = ?", *teamID)
	}
	if staffName != "" {
		query = query.Where("staff_snapshot LIKE ?", "%"+staffName+"%")
	}

	var result []models.ProShiftSnapshot
	if err := query.Find(&result).Error; err != nil {
		return nil, fmt.Errorf("查询快照失败: %w", err)
	}
	return result, nil
}

// RegenerateShiftSnapshot 手动重新生成指定日期/班次/设备的快照（先删旧再生成）
// Manually regenerate a snapshot for the given date/shift/device (deletes old, then recreates).
// 指定日・シフト・設備のスナップショットを手動で再生成する（旧レコード削除後に再作成）。
func (a *App) RegenerateShiftSnapshot(dateStr string, shiftID, deviceID int) error {
	if database.DB == nil {
		return fmt.Errorf("数据库未连接")
	}

	// 删旧
	database.DB.Where("snapshot_date = ? AND shift_id = ? AND device_id = ?",
		dateStr, shiftID, deviceID).Delete(&models.ProShiftSnapshot{})

	// 查班次配置
	var shiftModel models.SysShift
	if err := database.DB.Preload("Breaks").First(&shiftModel, shiftID).Error; err != nil {
		return fmt.Errorf("班次(id=%d)不存在: %w", shiftID, err)
	}
	sc := modelToShiftConfig(shiftModel)

	// 查设备
	var dev models.SysDevice
	if err := database.DB.First(&dev, deviceID).Error; err != nil {
		return fmt.Errorf("设备(id=%d)不存在: %w", deviceID, err)
	}

	baseDate, err := time.ParseInLocation("2006-01-02", dateStr, time.Now().Location())
	if err != nil {
		return fmt.Errorf("日期格式错误: %w", err)
	}
	shiftStart, shiftEnd := resolveShiftDatetime(sc, baseDate)

	globalCT := a.getGlobalCycleTimeFallback()
	ct := globalCT
	if dev.CycleTime != nil && *dev.CycleTime > 0 {
		ct = *dev.CycleTime
	}

	return generateOneSnapshot(dev, sc, shiftModel.ScheduleID, dateStr, shiftStart, shiftEnd, ct)
}

// BatchRegenerateResult 批量回填单条结果。
// Result item of a batch snapshot regeneration.
// バッチ再生成の1件分の結果。
type BatchRegenerateResult struct {
	Date     string `json:"date"`      // 快照日期 YYYY-MM-DD
	ShiftID  int    `json:"shift_id"`  // 班次 ID
	DeviceID int    `json:"device_id"` // 设备 ID
	OK       bool   `json:"ok"`        // 是否成功
	Error    string `json:"error"`     // 失败时的错误信息（成功时为空）
}

// BatchRegenerateShiftSnapshots 批量重新生成指定日期×班次×设备的快照。
// CN: 逐条调用单条重建逻辑（先删旧再聚合），单项失败不中断整批，返回每项结果清单。
//
//	dates 为日期列表（"YYYY-MM-DD"）；shiftIDs / deviceIDs 为过滤限定列表，空表示不过滤（即全部）。
//	典型用法：修复某几天快照时，传 dates=["2026-04-12","2026-04-13"]，shiftIDs/deviceIDs 传空。
//
// EN: Calls single-item regeneration for each combination; failures are logged per-item without aborting the batch.
//
//	Empty shiftIDs / deviceIDs means "all shifts / all devices" for the given dates.
//
// JP: 各組合せに対して単件再生成を呼び出し、失敗があっても他の件は継続する。
//
//	shiftIDs / deviceIDs が空の場合は該当日付の全班次・全設備を対象とする。
func (a *App) BatchRegenerateShiftSnapshots(dates []string, shiftIDs []int, deviceIDs []int) ([]BatchRegenerateResult, error) {
	if database.DB == nil {
		return nil, fmt.Errorf("数据库未连接")
	}
	if len(dates) == 0 {
		return nil, fmt.Errorf("dates 不能为空")
	}

	// 加载所有班次配置（避免在循环内反复查询）
	var allShifts []models.SysShift
	if err := database.DB.Preload("Breaks").Find(&allShifts).Error; err != nil {
		return nil, fmt.Errorf("查询班次配置失败: %w", err)
	}
	// 加载所有设备
	var allDevices []models.SysDevice
	if err := database.DB.Find(&allDevices).Error; err != nil {
		return nil, fmt.Errorf("查询设备列表失败: %w", err)
	}

	// 构建集合便于快速过滤
	shiftFilter := map[int]bool{}
	for _, id := range shiftIDs {
		shiftFilter[id] = true
	}
	deviceFilter := map[int]bool{}
	for _, id := range deviceIDs {
		deviceFilter[id] = true
	}

	globalCT := a.getGlobalCycleTimeFallback()

	var results []BatchRegenerateResult

	for _, dateStr := range dates {
		baseDate, err := time.ParseInLocation("2006-01-02", dateStr, time.Now().Location())
		if err != nil {
			results = append(results, BatchRegenerateResult{Date: dateStr, OK: false, Error: fmt.Sprintf("日期格式错误: %v", err)})
			continue
		}

		for _, shiftModel := range allShifts {
			if len(shiftFilter) > 0 && !shiftFilter[shiftModel.ID] {
				continue
			}
			sc := modelToShiftConfig(shiftModel)
			shiftStart, shiftEnd := resolveShiftDatetime(sc, baseDate)

			for _, dev := range allDevices {
				if len(deviceFilter) > 0 && !deviceFilter[dev.ID] {
					continue
				}

				ct := globalCT
				if dev.CycleTime != nil && *dev.CycleTime > 0 {
					ct = *dev.CycleTime
				}

				// 删旧记录
				database.DB.Where("snapshot_date = ? AND shift_id = ? AND device_id = ?",
					dateStr, shiftModel.ID, dev.ID).Delete(&models.ProShiftSnapshot{})

				// 重建
				genErr := generateOneSnapshot(dev, sc, shiftModel.ScheduleID, dateStr, shiftStart, shiftEnd, ct)
				item := BatchRegenerateResult{
					Date:     dateStr,
					ShiftID:  shiftModel.ID,
					DeviceID: dev.ID,
					OK:       genErr == nil,
				}
				if genErr != nil {
					item.Error = genErr.Error()
					log.Printf("[BatchRegenerateShiftSnapshots] 失败 date=%s shift=%d device=%d: %v",
						dateStr, shiftModel.ID, dev.ID, genErr)
				}
				results = append(results, item)
			}
		}
	}

	return results, nil
}

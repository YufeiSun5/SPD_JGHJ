package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// BreakTime 休息时间段。
// CN: 系统配置仍存放在用户目录 JSON，service 统一读写，避免 Wails 与 OEE 各自维护路径逻辑。
// EN: System config still lives in a user-directory JSON file; service centralizes file access for Wails and OEE.
// JP: システム設定はユーザーディレクトリの JSON に保存し、service で Wails/OEE のファイル処理を統一する。
type BreakTime struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	StartHour int    `json:"start_hour"`
	StartMin  int    `json:"start_min"`
	EndHour   int    `json:"end_hour"`
	EndMin    int    `json:"end_min"`
}

type UserConfig struct {
	ProductionCoefficient float64     `json:"production_coefficient"`
	DailyWorkMinutes      int         `json:"daily_work_minutes"`
	BreakTimes            []BreakTime `json:"break_times"`
}

func ConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("获取用户目录失败: %v", err)
	}
	configDir := filepath.Join(homeDir, ".spd_jghj")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("创建配置目录失败: %v", err)
	}
	return configDir, nil
}

func ConfigFilePath() (string, error) {
	configDir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "user_config.json"), nil
}

func DefaultBreakTimes() []BreakTime {
	return []BreakTime{
		{ID: 1, Name: "上午茶歇", StartHour: 9, StartMin: 40, EndHour: 9, EndMin: 50},
		{ID: 2, Name: "午餐休息", StartHour: 11, StartMin: 40, EndHour: 12, EndMin: 20},
		{ID: 3, Name: "下午茶歇", StartHour: 14, StartMin: 20, EndHour: 14, EndMin: 30},
	}
}

func DefaultConfig() *UserConfig {
	return &UserConfig{
		ProductionCoefficient: 100.0,
		DailyWorkMinutes:      460,
		BreakTimes:            DefaultBreakTimes(),
	}
}

func GetDailyWorkMinutes() (int, error) {
	config, err := GetSystemConfig()
	if err != nil {
		return 460, err
	}
	if config.DailyWorkMinutes <= 0 || config.DailyWorkMinutes > 1440 {
		return 460, nil
	}
	return config.DailyWorkMinutes, nil
}

func SetDailyWorkMinutes(minutes int) error {
	if minutes <= 0 || minutes > 1440 {
		return fmt.Errorf("每日工作分钟数必须在1-1440之间")
	}
	config, _ := GetSystemConfig()
	config.DailyWorkMinutes = minutes
	return SetSystemConfig(config)
}

func GetBreakTimes() ([]BreakTime, error) {
	config, err := GetSystemConfig()
	if err != nil {
		return DefaultBreakTimes(), err
	}
	if len(config.BreakTimes) == 0 {
		return DefaultBreakTimes(), nil
	}
	return config.BreakTimes, nil
}

func SetBreakTimes(breakTimes []BreakTime) error {
	if err := validateBreakTimes(breakTimes); err != nil {
		return err
	}
	config, _ := GetSystemConfig()
	config.BreakTimes = breakTimes
	return writeSystemConfig(config)
}

func GetSystemConfig() (*UserConfig, error) {
	configPath, err := ConfigFilePath()
	if err != nil {
		return DefaultConfig(), err
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		return DefaultConfig(), fmt.Errorf("读取配置文件失败: %v", err)
	}
	var config UserConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return DefaultConfig(), fmt.Errorf("解析配置文件失败: %v", err)
	}
	fillSystemConfigDefaults(&config)
	return &config, nil
}

func SetSystemConfig(config *UserConfig) error {
	if config.ProductionCoefficient <= 0 {
		return fmt.Errorf("单件加工时间必须大于0")
	}
	if config.DailyWorkMinutes <= 0 || config.DailyWorkMinutes > 1440 {
		return fmt.Errorf("每日工作分钟数必须在1-1440之间")
	}
	if err := validateBreakTimes(config.BreakTimes); err != nil {
		return err
	}
	return writeSystemConfig(config)
}

func fillSystemConfigDefaults(config *UserConfig) {
	if config.ProductionCoefficient <= 0 {
		config.ProductionCoefficient = 100.0
	}
	if config.DailyWorkMinutes <= 0 {
		config.DailyWorkMinutes = 460
	}
	if len(config.BreakTimes) == 0 {
		config.BreakTimes = DefaultBreakTimes()
	}
}

func validateBreakTimes(breakTimes []BreakTime) error {
	for _, bt := range breakTimes {
		if bt.StartHour < 0 || bt.StartHour > 23 || bt.EndHour < 0 || bt.EndHour > 23 {
			return fmt.Errorf("小时必须在0-23之间")
		}
		if bt.StartMin < 0 || bt.StartMin > 59 || bt.EndMin < 0 || bt.EndMin > 59 {
			return fmt.Errorf("分钟必须在0-59之间")
		}
		startInMin := bt.StartHour*60 + bt.StartMin
		endInMin := bt.EndHour*60 + bt.EndMin
		if startInMin >= endInMin {
			return fmt.Errorf("结束时间必须晚于开始时间")
		}
	}
	return nil
}

func writeSystemConfig(config *UserConfig) error {
	configPath, err := ConfigFilePath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("保存配置文件失败: %v", err)
	}
	return nil
}

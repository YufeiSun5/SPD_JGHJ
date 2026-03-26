package core

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
)

// InstanceLock 单实例锁
type InstanceLock struct {
	lockFile *os.File
	lockPath string
}

var globalInstanceLock *InstanceLock

// AcquireInstanceLock 获取单实例锁
// 确保同一台机器上只能运行一个实例
func AcquireInstanceLock(appName string) (*InstanceLock, error) {
	// 获取临时目录
	tempDir := os.TempDir()
	lockPath := filepath.Join(tempDir, fmt.Sprintf("%s.lock", appName))

	log.Printf("[SingleInstance] 锁文件路径: %s", lockPath)

	// 检查锁文件是否存在
	if fileExists(lockPath) {
		// 尝试读取PID并检查进程是否还在运行
		pid, err := readPIDFromFile(lockPath)
		if err == nil && isProcessRunning(pid) {
			return nil, fmt.Errorf("❌ 检测到已有实例正在运行 (PID: %d)\n"+
				"   锁文件: %s\n"+
				"   请先停止已运行的实例，或删除锁文件后重试", pid, lockPath)
		}

		// 如果进程不存在了，删除旧的锁文件
		log.Printf("[SingleInstance] 检测到遗留锁文件 (进程已停止)，正在清理...")
		os.Remove(lockPath)
	}

	// 创建锁文件 (独占模式)
	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
	if err != nil {
		if os.IsExist(err) {
			// 如果文件已存在 (多个实例同时启动的竞争情况)
			return nil, fmt.Errorf("❌ 另一个实例正在同时启动，锁文件已被占用: %s", lockPath)
		}
		return nil, fmt.Errorf("创建锁文件失败: %w", err)
	}

	// 写入当前进程的PID
	currentPID := os.Getpid()
	if _, err := lockFile.WriteString(fmt.Sprintf("%d\n", currentPID)); err != nil {
		lockFile.Close()
		os.Remove(lockPath)
		return nil, fmt.Errorf("写入PID失败: %w", err)
	}

	lockFile.Sync()

	lock := &InstanceLock{
		lockFile: lockFile,
		lockPath: lockPath,
	}

	globalInstanceLock = lock
	log.Printf("[SingleInstance] ✅ 单实例锁已获取 (PID: %d)", currentPID)

	return lock, nil
}

// Release 释放锁
func (lock *InstanceLock) Release() error {
	if lock == nil || lock.lockFile == nil {
		return nil
	}

	log.Printf("[SingleInstance] 释放单实例锁: %s", lock.lockPath)

	// 关闭文件
	if err := lock.lockFile.Close(); err != nil {
		log.Printf("[SingleInstance] 关闭锁文件失败: %v", err)
	}

	// 删除锁文件
	if err := os.Remove(lock.lockPath); err != nil {
		log.Printf("[SingleInstance] 删除锁文件失败: %v", err)
		return err
	}

	log.Println("[SingleInstance] ✅ 单实例锁已释放")
	return nil
}

// ReleaseGlobalLock 释放全局锁 (用于优雅退出)
func ReleaseGlobalLock() {
	if globalInstanceLock != nil {
		globalInstanceLock.Release()
	}
}

// fileExists 检查文件是否存在
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// readPIDFromFile 从锁文件读取PID
func readPIDFromFile(path string) (int, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return 0, err
	}

	pidStr := string(data)
	pid, err := strconv.Atoi(pidStr[:len(pidStr)-1]) // 去掉换行符
	if err != nil {
		return 0, err
	}

	return pid, nil
}

// isProcessRunning 检查进程是否正在运行 (Windows)
func isProcessRunning(pid int) bool {
	// 尝试获取进程句柄
	handle, err := syscall.OpenProcess(syscall.PROCESS_QUERY_INFORMATION, false, uint32(pid))
	if err != nil {
		// 如果无法打开进程，说明进程不存在
		return false
	}
	syscall.CloseHandle(handle)
	return true
}

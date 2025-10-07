/*
 * @Descripttion:
 * @version:
 * @Author: lfzxs@qq.com
 * @Date: 2025-10-07 21:56:40
 * @LastEditors: lfzxs@qq.com
 * @LastEditTime: 2025-10-07 22:29:13
 */
package sysinfo

import (
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

// Info 存储系统信息
type Info struct {
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
}

// GetInfo 获取当前的系统信息（CPU 和内存使用率）
func GetInfo() (*Info, error) {
	// 获取内存使用率
	vm, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	// 获取CPU使用率
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return nil, err
	}

	info := &Info{
		CPUUsage:    cpuPercent[0],
		MemoryUsage: vm.UsedPercent,
	}

	return info, nil
}

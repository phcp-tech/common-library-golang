// Copyright(C) 2019-2026 PHCP Technologies. All rights reserved.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// 	http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build linux
// +build linux

package cgroup

import (
	"bufio"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/phcp-tech/common-library-golang/log"
)

// CPULimitMilli 返回当前进程 cgroup 的 CPU 限制（毫核，mCPU）。
// 不限时返回 0。
func CPULimitMilli() (int, error) {
	// 尝试 cgroup v2
	if v2Root, v2Path, ok := detectCgroupV2(); ok {
		dir := filepath.Join(v2Root, v2Path)
		data, err := os.ReadFile(filepath.Join(dir, "cpu.max"))
		if err == nil {
			fields := strings.Fields(string(data))
			if len(fields) >= 2 {
				if fields[0] == "max" {
					log.Debug("cgroup v2 cpu limit: unlimited")
					return 0, nil // unlimited
				}
				quota, err1 := parseInt(fields[0])
				period, err2 := parseInt(fields[1])
				if err1 == nil && err2 == nil && period > 0 {
					// limit cores = quota/period；转毫核
					log.Debugf("cgroup v2 cpu limit: %d mCPU", int((float64(quota)/float64(period))*1000.0))
					return int((float64(quota) / float64(period)) * 1000.0), nil
				}
			}
		}
		// 读取失败则继续回退 v1
	}

	// 回退到 cgroup v1
	v1Dir, err := cgroupV1CpuDir()
	if err != nil {
		return 0, fmt.Errorf("detect cgroup v1 cpu dir: %w", err)
	}
	quotaB, err1 := os.ReadFile(filepath.Join(v1Dir, "cpu.cfs_quota_us"))
	periodB, err2 := os.ReadFile(filepath.Join(v1Dir, "cpu.cfs_period_us"))
	if err1 != nil || err2 != nil {
		return 0, fmt.Errorf("read v1 quota/period: %v %v", err1, err2)
	}
	quota, _ := parseInt(strings.TrimSpace(string(quotaB)))
	period, _ := parseInt(strings.TrimSpace(string(periodB)))
	if quota <= 0 || period <= 0 || quota == -1 {
		log.Info("cgroup v1 cpu limit: unlimited")
		return 0, nil // unlimited 或未设置
	}
	log.Infof("cgroup v1 cpu limit: %d mCPU", int((float64(quota)/float64(period))*1000.0))
	return int((float64(quota) / float64(period)) * 1000.0), nil
}

// CPURequestMilli 返回当前进程 cgroup 的 CPU 请求（毫核，mCPU）。
// 未显式请求时返回 0。
func CPURequestMilli() (int, error) {
	// 先 v2：通过 cpu.weight -> 还原到 v1 shares -> 换算毫核
	if v2Root, v2Path, ok := detectCgroupV2(); ok {
		dir := filepath.Join(v2Root, v2Path)
		b, err := os.ReadFile(filepath.Join(dir, "cpu.weight"))
		if err == nil {
			w64, err := parseInt(strings.TrimSpace(string(b)))
			if err == nil {
				weight := int64(w64)
				if weight < 1 {
					weight = 1
				}
				if weight > 10000 {
					weight = 10000
				}
				// weight -> mCPU. The runtime converts shares to weight with a new formula (1024 shares = weight 100).
				// The old linear formula showed a 500m request as 1511m, see cpuWeightToMilli.
				// weight 1 = shares 2 = BestEffort, treated as no request.
				if weight <= 1 {
					log.Debugf("cgroup v2 cpu request: weight <= 1, treated as no request")
					return 0, nil
				}
				mcpu := cpuWeightToMilli(weight)
				log.Debugf("cgroup v2 cpu request: %d mCPU", mcpu)
				return mcpu, nil
			}
		}
		// 失败则回退 v1
	}

	// v1：cpu.shares -> mCPU
	v1Dir, err := cgroupV1CpuDir()
	if err != nil {
		return 0, fmt.Errorf("detect cgroup v1 cpu dir: %w", err)
	}
	b, err := os.ReadFile(filepath.Join(v1Dir, "cpu.shares"))
	if err != nil {
		return 0, fmt.Errorf("read cpu.shares: %w", err)
	}
	shares, err := parseInt(strings.TrimSpace(string(b)))
	if err != nil {
		return 0, err
	}
	if shares <= 2 {
		return 0, nil
	}
	mcpu := int((int64(shares)*1000 + 512) / 1024)
	log.Debugf("cgroup v1 cpu request: %d mCPU", mcpu)
	return mcpu, nil
}

// ---------- 辅助：检测 cgroup v2 / 解析 v1 mount & path ----------

func detectCgroupV2() (root, path string, ok bool) {
	root = "/sys/fs/cgroup"
	if _, err := os.Stat(filepath.Join(root, "cgroup.controllers")); err != nil {
		return "", "", false
	}
	f, err := os.Open("/proc/self/cgroup")
	if err != nil {
		return "", "", false
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		// v2 行形如：0::/docker/<id> 或 0::/
		line := sc.Text()
		parts := strings.SplitN(line, ":", 3)
		if len(parts) == 3 && parts[0] == "0" && parts[1] == "" {
			return root, parts[2], true
		}
	}
	return "", "", false
}

func cgroupV1CpuDir() (string, error) {
	// 1) 从 /proc/self/cgroup 拿到 cpu 的相对路径
	var rel string
	{
		f, err := os.Open("/proc/self/cgroup")
		if err != nil {
			return "", err
		}
		defer f.Close()
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			// 形如：2:cpu,cpuacct:/kubepods.slice/...
			line := sc.Text()
			parts := strings.SplitN(line, ":", 3)
			if len(parts) != 3 {
				continue
			}
			controllers := strings.Split(parts[1], ",")
			for _, c := range controllers {
				if c == "cpu" { // 选到 cpu 这一行
					rel = parts[2]
					break
				}
			}
			if rel != "" {
				break
			}
		}
		if rel == "" {
			return "", errors.New("no cpu controller entry in /proc/self/cgroup")
		}
	}

	// 2) 从 /proc/self/mountinfo 找到 cpu 控制器的挂载点
	mnt, err := findCgroupV1Mountpoint("cpu")
	if err != nil {
		return "", err
	}
	return filepath.Join(mnt, rel), nil
}

func findCgroupV1Mountpoint(controller string) (string, error) {
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return "", err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		// mountinfo 最后三列形如："- cgroup cgroup rw,relatime,cpu,cpuacct"
		line := sc.Text()
		// 拆 fstype 与 super options
		parts := strings.Split(line, " - ")
		if len(parts) != 2 {
			continue
		}
		left := strings.Fields(parts[0])
		right := strings.Fields(parts[1])
		if len(left) < 5 || len(right) < 3 {
			continue
		}
		mountPoint := left[4]
		fsType := right[0]
		superOpts := right[2]
		if fsType == "cgroup" {
			for _, opt := range strings.Split(superOpts, ",") {
				if opt == controller {
					return mountPoint, nil
				}
			}
		}
	}
	return "", fmt.Errorf("cgroup v1 mountpoint for %q not found", controller)
}

// MemoryLimitBytes 返回当前进程 cgroup 的内存限制（字节）。
// 不限时返回 0。
func MemoryLimitBytes() (int64, error) {
	// 尝试 cgroup v2
	if v2Root, v2Path, ok := detectCgroupV2(); ok {
		dir := filepath.Join(v2Root, v2Path)
		data, err := os.ReadFile(filepath.Join(dir, "memory.max"))
		if err == nil {
			limitStr := strings.TrimSpace(string(data))
			if limitStr == "max" {
				log.Debug("cgroup v2 memory limit: unlimited")
				return 0, nil // unlimited
			}
			limit, err := strconv.ParseInt(limitStr, 10, 64)
			if err == nil {
				log.Debugf("cgroup v2 memory limit: %d bytes", limit)
				return limit, nil
			}
		}
		// 读取失败则继续回退 v1
	}

	// 回退到 cgroup v1
	v1Dir, err := cgroupV1MemoryDir()
	if err != nil {
		return 0, fmt.Errorf("detect cgroup v1 memory dir: %w", err)
	}
	limitB, err := os.ReadFile(filepath.Join(v1Dir, "memory.limit_in_bytes"))
	if err != nil {
		return 0, fmt.Errorf("read memory.limit_in_bytes: %w", err)
	}
	limitStr := strings.TrimSpace(string(limitB))
	limit, err := strconv.ParseInt(limitStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse memory limit: %w", err)
	}
	// 检查是否是一个非常大的值，表示无限制
	// 通常在cgroup v1中，无限制时会设置为一个极大值，如 math.MaxInt64
	if limit == math.MaxInt64 || limit > (1<<62) {
		log.Debug("cgroup v1 memory limit: unlimited (very large value)")
		return 0, nil
	}
	log.Debugf("cgroup v1 memory limit: %d bytes", limit)
	return limit, nil
}

// MemoryRequestBytes 返回当前进程 cgroup 的内存请求（字节）。
// 未显式请求时返回 0。
//
// 注意：在 Kubernetes 中，容器的 memory request 默认不会写入 cgroup，只用于调度和 OOM 打分。
// 本函数读取的 memory.low（v2）和 memory.soft_limit_in_bytes（v1），K8s 都不设置。
// 只有开启 MemoryQoS（cgroup v2）时，kubelet 才会把 request 写入 memory.min，而这里没有读取 memory.min。
// 所以 0 不表示"没有设置 request"，而表示"无法获取"。已在 UAT 容器内验证：Octopus 设置了 request，
// 但 memory.min 和 memory.low 都是 0。metrics.GetMetrics 因此不再输出 memoryRequest，不要用本函数的结果判断 request。
func MemoryRequestBytes() (int64, error) {
	// 对于内存请求，在 Kubernetes 中通常通过 memory.soft_limit_in_bytes (v1)
	// 或者其他机制来实现，但这不是标准的 cgroup 特性
	// 在实际环境中，memory request 通常不会直接反映在 cgroup 中

	// 尝试 cgroup v2 - 检查 memory.low (软限制，类似于 request)
	if v2Root, v2Path, ok := detectCgroupV2(); ok {
		dir := filepath.Join(v2Root, v2Path)
		data, err := os.ReadFile(filepath.Join(dir, "memory.low"))
		if err == nil {
			lowStr := strings.TrimSpace(string(data))
			if lowStr == "max" || lowStr == "0" {
				log.Debug("cgroup v2 memory request: no soft limit set")
				return 0, nil
			}
			low, err := strconv.ParseInt(lowStr, 10, 64)
			if err == nil && low > 0 {
				log.Debugf("cgroup v2 memory request (soft limit): %d bytes", low)
				return low, nil
			}
		}
		// 读取失败则继续回退 v1
	}

	// 回退到 cgroup v1 - 检查 memory.soft_limit_in_bytes
	v1Dir, err := cgroupV1MemoryDir()
	if err != nil {
		return 0, fmt.Errorf("detect cgroup v1 memory dir: %w", err)
	}
	softLimitB, err := os.ReadFile(filepath.Join(v1Dir, "memory.soft_limit_in_bytes"))
	if err != nil {
		// 如果没有 soft_limit 文件，返回 0 表示没有设置请求
		log.Debug("cgroup v1 memory request: memory.soft_limit_in_bytes not found")
		return 0, nil
	}
	softLimitStr := strings.TrimSpace(string(softLimitB))
	softLimit, err := strconv.ParseInt(softLimitStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse memory soft limit: %w", err)
	}
	// 检查是否是一个非常大的值，表示无限制
	if softLimit == math.MaxInt64 || softLimit > (1<<62) {
		log.Debug("cgroup v1 memory request: unlimited (very large soft limit)")
		return 0, nil
	}
	log.Debugf("cgroup v1 memory request (soft limit): %d bytes", softLimit)
	return softLimit, nil
}

// cgroupV1MemoryDir 获取 cgroup v1 memory 控制器的目录路径
func cgroupV1MemoryDir() (string, error) {
	// 1) 从 /proc/self/cgroup 拿到 memory 的相对路径
	var rel string
	{
		f, err := os.Open("/proc/self/cgroup")
		if err != nil {
			return "", err
		}
		defer f.Close()
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			// 形如：3:memory:/kubepods.slice/...
			line := sc.Text()
			parts := strings.SplitN(line, ":", 3)
			if len(parts) != 3 {
				continue
			}
			controllers := strings.Split(parts[1], ",")
			for _, c := range controllers {
				if c == "memory" { // 选到 memory 这一行
					rel = parts[2]
					break
				}
			}
			if rel != "" {
				break
			}
		}
		if rel == "" {
			return "", errors.New("no memory controller entry in /proc/self/cgroup")
		}
	}

	// 2) 从 /proc/self/mountinfo 找到 memory 控制器的挂载点
	mnt, err := findCgroupV1Mountpoint("memory")
	if err != nil {
		return "", err
	}
	return filepath.Join(mnt, rel), nil
}

func parseInt(s string) (int, error) {
	s = strings.TrimSpace(s)
	return strconv.Atoi(s)
}

// CPUThrottleStat 返回当前进程 cgroup 累计的 CPU 节流计数（cpu.stat）。
// 计数只增不减，调用方需要取两次读数的差值。
// 不限 CPU 时各字段为 0。读取失败时返回错误。
func CPUThrottleStat() (CPUThrottle, error) {
	// 先 v2
	if v2Root, v2Path, ok := detectCgroupV2(); ok {
		data, err := os.ReadFile(filepath.Join(v2Root, v2Path, "cpu.stat"))
		if err == nil {
			return parseCPUStat(string(data)), nil
		}
		// 读取失败则继续回退 v1
	}

	// 回退到 cgroup v1
	v1Dir, err := cgroupV1CpuDir()
	if err != nil {
		return CPUThrottle{}, fmt.Errorf("detect cgroup v1 cpu dir: %w", err)
	}
	data, err := os.ReadFile(filepath.Join(v1Dir, "cpu.stat"))
	if err != nil {
		return CPUThrottle{}, fmt.Errorf("read cpu.stat: %w", err)
	}
	return parseCPUStat(string(data)), nil
}

package config

import (
	"os"
	"path/filepath"
)

// findFileUpwards 从起始目录向上查找指定文件，返回第一个匹配的绝对路径
func findFileUpwards(startDir, filename string) (string, bool) {
	dir := startDir
	for {
		candidate := filepath.Join(dir, filename)
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", false
}

// FindFileUpwardsProxy 导出供其他包使用的向上查找函数
func FindFileUpwardsProxy(startDir, filename string) (string, bool) {
	return findFileUpwards(startDir, filename)
}

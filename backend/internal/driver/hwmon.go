package driver

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var allowedPathPrefixes = []string{
	"/sys/class/hwmon/",
	"/sys/devices/",
}

// ValidateHwmonPath 检查路径是否在允许的 sysfs 目录下，防止任意文件读写。
func ValidateHwmonPath(path string) error {
	if path == "" {
		return nil
	}
	cleaned := filepath.Clean(path)
	if strings.Contains(cleaned, "..") {
		return fmt.Errorf("路径不允许包含 '..'：%s", path)
	}
	for _, prefix := range allowedPathPrefixes {
		if strings.HasPrefix(cleaned, prefix) {
			return nil
		}
	}
	return fmt.Errorf("路径不在允许范围内（需位于 /sys/class/hwmon/ 或 /sys/devices/ 下）：%s", path)
}

type HWMONDriver struct{}

func NewHWMONDriver() *HWMONDriver {
	return &HWMONDriver{}
}

func (d *HWMONDriver) ReadRPM(path string) (int, error) {
	if err := ValidateHwmonPath(path); err != nil {
		return 0, err
	}
	return readIntFile(path)
}

func (d *HWMONDriver) ReadPWM(path string) (int, error) {
	if err := ValidateHwmonPath(path); err != nil {
		return 0, err
	}
	return readIntFile(path)
}

func (d *HWMONDriver) WritePWM(enablePath, pwmPath string, pwm int) error {
	if err := ValidateHwmonPath(enablePath); err != nil {
		return err
	}
	if err := ValidateHwmonPath(pwmPath); err != nil {
		return err
	}
	if enablePath != "" {
		if err := os.WriteFile(enablePath, []byte("1\n"), 0o644); err != nil {
			return fmt.Errorf("写入 PWM 使能路径失败：%w", err)
		}
	}
	return os.WriteFile(pwmPath, []byte(fmt.Sprintf("%d\n", pwm)), 0o644)
}

func (d *HWMONDriver) ScanFans() ([]map[string]string, error) {
	entries, err := filepath.Glob("/sys/class/hwmon/hwmon*")
	if err != nil {
		return nil, err
	}
	// 使用非 nil 空切片，JSON 序列化为 []；nil 切片会变成 null 导致前端误判
	fans := make([]map[string]string, 0)
	for _, hwmon := range entries {
		pwmFiles, _ := filepath.Glob(filepath.Join(hwmon, "pwm[0-9]"))
		for _, pwmFile := range pwmFiles {
			index := strings.TrimPrefix(filepath.Base(pwmFile), "pwm")
			rpmPath := filepath.Join(hwmon, "fan"+index+"_input")
			if _, err := os.Stat(rpmPath); err != nil {
				rpmPath = ""
			}
			enablePath := filepath.Join(hwmon, "pwm"+index+"_enable")
			if _, err := os.Stat(enablePath); err != nil {
				enablePath = ""
			}
			nameFile := filepath.Join(hwmon, "name")
			name := filepath.Base(hwmon)
			if content, err := os.ReadFile(nameFile); err == nil {
				name = strings.TrimSpace(string(content))
			}
			fans = append(fans, map[string]string{
				"id":          fmt.Sprintf("%s-pwm%s", filepath.Base(hwmon), index),
				"name":        fmt.Sprintf("%s 风扇 %s", name, index),
				"pwm_path":    pwmFile,
				"rpm_path":    rpmPath,
				"enable_path": enablePath,
			})
		}
	}
	return fans, nil
}

func readIntFile(path string) (int, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(string(raw)))
}

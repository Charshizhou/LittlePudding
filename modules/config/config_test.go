package config

import (
	"fmt"
	"testing"
)

func TestInitConfig(t *testing.T) {
	var s *Setting
	s, err := ReadConfig("config.ini")
	if err != nil {
		t.Fatalf("读取配置文件失败-%s", err)
	}
	fmt.Println(s)
}

func TestInstallConfig(t *testing.T) {
	s := DefaultSetting
	err := WriteConfig("config.ini", &s)
	if err != nil {
		t.Fatalf("写入配置文件失败-%s", err)
	}
}

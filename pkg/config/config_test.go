package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	yamlContent := `
default:
  inputs: ["./data/*.log"]
  output: "./out.log"
server:
  enabled: true
  port: 8080
`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("임시 설정 파일 생성 실패: %v", err)
	}

	cfg, err := LoadConfig(configPath)

	if err != nil {
		t.Fatalf("LoadConfig() 에러 발생: %v", err)
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("포트 설정 불일치: 예상 8080, 결과 %d", cfg.Server.Port)
	}

	if len(cfg.Default.Inputs) != 1 || cfg.Default.Inputs[0] != "./data/*.log" {
		t.Errorf("입력 경로 설정 불일치")
	}

	_, err = LoadConfig("non_existent.yaml")
	if err == nil {
		t.Error("존재하지 않는 파일 로드 시 에러가 발생해야 합니다.")
	}
}

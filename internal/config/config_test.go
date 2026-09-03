package config

import (
	"strings"
	"testing"
)

func TestGetDecryptedDSN_PasswordCommand(t *testing.T) {
	tests := []struct {
		name         string
		conn         SQLConfig
		wantContains string
		wantMissing  string
	}{
		{
			name: "command output used as password for field-based connection",
			conn: SQLConfig{
				Driver:          "postgres",
				Host:            "localhost",
				Port:            5432,
				Database:        "mydb",
				Username:        "alice",
				PasswordCommand: "echo secretpass",
			},
			wantContains: "alice:secretpass@",
		},
		{
			name: "command output injected into explicit DSN",
			conn: SQLConfig{
				PasswordCommand: "echo secretpass",
				DSN:             "postgres://alice@localhost:5432/mydb?sslmode=disable",
			},
			wantContains: "alice:secretpass@",
		},
		{
			name: "failed command returns DSN without password",
			conn: SQLConfig{
				Driver:          "postgres",
				Host:            "localhost",
				Port:            5432,
				Database:        "mydb",
				Username:        "alice",
				PasswordCommand: "false",
			},
			wantContains: "alice@",
			wantMissing:  "alice:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dsn := tt.conn.GetDecryptedDSN()
			if tt.wantContains != "" && !strings.Contains(dsn, tt.wantContains) {
				t.Errorf("DSN %q does not contain %q", dsn, tt.wantContains)
			}
			if tt.wantMissing != "" && strings.Contains(dsn, tt.wantMissing) {
				t.Errorf("DSN %q should not contain %q", dsn, tt.wantMissing)
			}
		})
	}
}

func TestIsPasswordReadable_PasswordCommand(t *testing.T) {
	conn := SQLConfig{PasswordCommand: "echo secret"}
	if !conn.IsPasswordReadable() {
		t.Error("IsPasswordReadable should return true when PasswordCommand is set")
	}
}

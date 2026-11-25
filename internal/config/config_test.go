package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolvePassword(t *testing.T) {
	tests := []struct {
		name    string
		conn    SQLConfig
		want    string
		wantErr bool
	}{
		{
			name: "password command output is used",
			conn: SQLConfig{PasswordCommand: "echo secretpass"},
			want: "secretpass",
		},
		{
			name: "command output is trimmed of surrounding whitespace",
			conn: SQLConfig{PasswordCommand: "printf '  spaced  \\n'"},
			want: "spaced",
		},
		{
			name:    "failed password command errors",
			conn:    SQLConfig{PasswordCommand: "false"},
			wantErr: true,
		},
		{
			name: "plaintext password passes through",
			conn: SQLConfig{Password: "plain"},
			want: "plain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.conn.ResolvePassword()
			if (err != nil) != tt.wantErr {
				t.Fatalf("ResolvePassword() err = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ResolvePassword() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestValidatePasswordCommand(t *testing.T) {
	tests := []struct {
		name    string
		conn    SQLConfig
		wantErr bool
	}{
		{
			name: "field mode with command is allowed",
			conn: SQLConfig{Host: "localhost", PasswordCommand: "echo x"},
		},
		{
			name: "field mode with masked display DSN is allowed",
			conn: SQLConfig{Host: "localhost", DSN: "postgres://a:***@h/db", PasswordCommand: "echo x"},
		},
		{
			name: "raw DSN without command is allowed",
			conn: SQLConfig{DSN: "postgres://a@h/db"},
		},
		{
			name:    "raw DSN with command is rejected",
			conn:    SQLConfig{DSN: "postgres://a@h/db", PasswordCommand: "echo x"},
			wantErr: true,
		},
		{
			name:    "plain password with command is rejected",
			conn:    SQLConfig{Host: "localhost", Password: "plain", PasswordCommand: "echo x"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.conn.ValidatePasswordCommand(); (err != nil) != tt.wantErr {
				t.Errorf("ValidatePasswordCommand() err = %v, wantErr %v", err, tt.wantErr)
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

func TestPasswordCommandIsCached(t *testing.T) {
	runs := filepath.Join(t.TempDir(), "runs")
	conn := SQLConfig{PasswordCommand: "echo run >> " + runs + "; echo cachedpass"}
	t.Cleanup(conn.ForgetPassword)

	for range 3 {
		got, err := conn.ResolvePassword()
		if err != nil {
			t.Fatalf("ResolvePassword() err = %v", err)
		}
		if got != "cachedpass" {
			t.Fatalf("ResolvePassword() = %q, want cachedpass", got)
		}
	}
	if count := countLines(t, runs); count != 1 {
		t.Errorf("password command ran %d times, want 1", count)
	}

	conn.ForgetPassword()
	if _, err := conn.ResolvePassword(); err != nil {
		t.Fatalf("ResolvePassword() after ForgetPassword err = %v", err)
	}
	if count := countLines(t, runs); count != 2 {
		t.Errorf("password command ran %d times after ForgetPassword, want 2", count)
	}
}

func TestFailedPasswordCommandIsNotCached(t *testing.T) {
	ready := filepath.Join(t.TempDir(), "ready")
	conn := SQLConfig{PasswordCommand: "test -f " + ready + " && echo laterpass"}
	t.Cleanup(conn.ForgetPassword)

	if _, err := conn.ResolvePassword(); err == nil {
		t.Fatal("ResolvePassword() succeeded before the command could, want an error")
	}
	if err := os.WriteFile(ready, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := conn.ResolvePassword()
	if err != nil {
		t.Fatalf("ResolvePassword() err = %v, want the failure not to be cached", err)
	}
	if got != "laterpass" {
		t.Errorf("ResolvePassword() = %q, want laterpass", got)
	}
}

func countLines(t *testing.T, path string) int {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return strings.Count(string(content), "\n")
}

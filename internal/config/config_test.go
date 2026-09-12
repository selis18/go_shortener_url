package config

import (
	"flag"
	"os"
	"testing"
)

func TestDatabaseDSN(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		env  string
		want string
	}{
		{"default", nil, "", ""},
		{"flag", []string{"-d", "postgres://flag/db"}, "", "postgres://flag/db"},
		{"environment", nil, "postgres://env/db", "postgres://env/db"},
		{"environment overrides flag", []string{"-d", "postgres://flag/db"}, "postgres://env/db", "postgres://env/db"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			oldFlags, oldArgs := flag.CommandLine, os.Args
			oldAddress, oldHost, oldLevel, oldPath, oldDSN := address, host, level, filePath, databaseDSN
			t.Cleanup(func() {
				flag.CommandLine, os.Args = oldFlags, oldArgs
				address, host, level, filePath, databaseDSN = oldAddress, oldHost, oldLevel, oldPath, oldDSN
			})
			for _, key := range []string{"SERVER_ADDRESS", "BASE_URL", "LOG_LEVEL", "FILE_STORAGE_PATH"} {
				t.Setenv(key, "")
			}
			t.Setenv("DATABASE_DSN", tc.env)
			flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
			os.Args = append([]string{"test"}, tc.args...)
			ParseConfig()
			if got := GetDatabaseDSN(); got != tc.want {
				t.Fatalf("DSN = %q, want %q", got, tc.want)
			}
		})
	}
}

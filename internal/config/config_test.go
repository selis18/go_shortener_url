package config

import (
	"flag"
	"os"
	"testing"
	"time"
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
			oldBatch, oldInterval, oldSecret := deleteBatchSize, deleteFlushInterval, secretKey
			t.Cleanup(func() {
				flag.CommandLine, os.Args = oldFlags, oldArgs
				address, host, level, filePath, databaseDSN = oldAddress, oldHost, oldLevel, oldPath, oldDSN
				deleteBatchSize, deleteFlushInterval, secretKey = oldBatch, oldInterval, oldSecret
			})
			for _, key := range []string{"SERVER_ADDRESS", "BASE_URL", "LOG_LEVEL", "FILE_STORAGE_PATH", "SECRET_KEY", "DELETE_BATCH_SIZE", "DELETE_FLUSH_INTERVAL"} {
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

func TestDeleteConfig(t *testing.T) {
	for _, tc := range []struct {
		name                  string
		args                  []string
		batchEnv, intervalEnv string
		wantBatch             int
		wantInterval          time.Duration
	}{
		{"defaults", nil, "", "", 1000, 25 * time.Millisecond},
		{"flags", []string{"-delete-batch-size", "20", "-delete-flush-interval", "2s"}, "", "", 20, 2 * time.Second},
		{"environment", nil, "50", "100ms", 50, 100 * time.Millisecond},
		{"environment overrides flags", []string{"-delete-batch-size", "20", "-delete-flush-interval", "2s"}, "50", "100ms", 50, 100 * time.Millisecond},
	} {
		t.Run(tc.name, func(t *testing.T) {
			oldFlags, oldArgs := flag.CommandLine, os.Args
			oldAddress, oldHost, oldLevel, oldPath, oldDSN := address, host, level, filePath, databaseDSN
			oldBatch, oldInterval, oldSecret := deleteBatchSize, deleteFlushInterval, secretKey
			t.Cleanup(func() {
				flag.CommandLine, os.Args = oldFlags, oldArgs
				address, host, level, filePath, databaseDSN = oldAddress, oldHost, oldLevel, oldPath, oldDSN
				deleteBatchSize, deleteFlushInterval, secretKey = oldBatch, oldInterval, oldSecret
			})
			for _, key := range []string{"SERVER_ADDRESS", "BASE_URL", "LOG_LEVEL", "FILE_STORAGE_PATH", "DATABASE_DSN", "SECRET_KEY"} {
				t.Setenv(key, "")
			}
			t.Setenv("DELETE_BATCH_SIZE", tc.batchEnv)
			t.Setenv("DELETE_FLUSH_INTERVAL", tc.intervalEnv)
			flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
			os.Args = append([]string{"test"}, tc.args...)
			ParseConfig()
			if GetDeleteBatchSize() != tc.wantBatch || GetDeleteFlushInterval() != tc.wantInterval {
				t.Fatalf("delete config = (%d, %s), want (%d, %s)", GetDeleteBatchSize(), GetDeleteFlushInterval(), tc.wantBatch, tc.wantInterval)
			}
		})
	}
}

package e2e

import (
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	// Wait a bit for services to be ready (in case docker-compose is run manually)
	time.Sleep(5 * time.Second)

	code := m.Run()

	os.Exit(code)
}

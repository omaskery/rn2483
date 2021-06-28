package testutils

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"testing"

	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
)

// CreateTestLogger creates a logr.Logger for tests which optionally (based on environment variables)
// can capture the log output and then only print it on test failure
func CreateTestLogger(t *testing.T) logr.Logger {
	var destination io.Writer = os.Stdout

	captureLogs := os.Getenv("CAPTURE_LOGS")
	if captureLogs == "" || captureLogs == "1" {
		buffer := &bytes.Buffer{}
		t.Cleanup(func() {
			if t.Failed() {
				fmt.Println("dumping log output for failed test:")
				fmt.Print(buffer)
			}
		})
		destination = buffer
	}

	logger := stdr.New(log.New(destination, "", log.LstdFlags | log.Lmicroseconds)).WithName(t.Name())

	return logger
}

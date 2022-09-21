//go:build e2e

package test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/JenswBE/dead-link-checker/cmd/config"
	"github.com/JenswBE/dead-link-checker/internal"
	"github.com/JenswBE/dead-link-checker/internal/report"
)

func TestE2E(t *testing.T) {
	// Parse config
	delicConfig := mustV(config.ParseConfig("./config.yml"))

	// Check website
	manager := internal.NewManager()
	actual := manager.Run(delicConfig)

	// Load expected
	expectedFile := mustV(os.Open("./expected.jsonc"))
	defer expectedFile.Close()
	expectedFileWarning := "// This file used for test assertion. Should be manually rechecked on any changes."
	header := make([]byte, len(expectedFileWarning))
	_ = mustV(expectedFile.Read(header))
	if !strings.HasPrefix(string(header), expectedFileWarning) {
		t.Fatal("expected.json is regenerated without manually rechecking.")
	}
	var expected map[string]report.Report
	must(json.NewDecoder(expectedFile).Decode(&expected))

	// Assert
	diff := cmp.Diff(expected, actual)
	if diff != "" {
		t.Fatal(diff)
	}
}

// must panics if provided error is not nil.
func must(err error) {
	_ = mustV("", err)
}

// mustV works same as "must", but also returns the value if error is nil.
func mustV[T any](v T, err error) T {
	if err != nil {
		panic(err.Error())
	}
	return v
}

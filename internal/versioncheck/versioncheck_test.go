package versioncheck

import "testing"

func TestResultHasUpdateSkipsDevBuilds(t *testing.T) {
	if compare("9.9.9", "dev").HasUpdate() {
		t.Fatal("dev builds should not report updates")
	}
	if !compare("9.9.9", "1.0.0").HasUpdate() {
		t.Fatal("older release builds should report updates")
	}
}

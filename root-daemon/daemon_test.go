package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"free-mind/ipc"
)

// setupTestPaths points the daemon's file path functions at temp files
// and returns a cleanup function that resets them.
func setupTestPaths(t *testing.T) (hostsPath, sitesPath string) {
	t.Helper()
	dir := t.TempDir()

	hostsPath = filepath.Join(dir, "hosts")
	sitesPath = filepath.Join(dir, "hosts-list-to-be-blocked")

	hostsFilePathOverride = hostsPath
	sitesToBlockPathOverride = sitesPath

	t.Cleanup(func() {
		hostsFilePathOverride = ""
		sitesToBlockPathOverride = ""
	})
	return hostsPath, sitesPath
}

// TestUpdateSitesToBeBlocked verifies that a comma-separated site list is written
// as "127.0.0.1 <site>" lines.
func TestUpdateSitesToBeBlocked(t *testing.T) {
	_, sitesPath := setupTestPaths(t)

	UpdateSitesToBeBlocked("reddit.com, twitter.com, facebook.com")

	data, err := os.ReadFile(sitesPath)
	if err != nil {
		t.Fatalf("reading sites file: %v", err)
	}

	got := string(data)
	for _, site := range []string{"reddit.com", "twitter.com", "facebook.com"} {
		want := "127.0.0.1 " + site
		if !strings.Contains(got, want) {
			t.Errorf("sites file missing %q\ngot:\n%s", want, got)
		}
	}
}

// TestUpdateSitesToBeBlocked_SkipsBlanks verifies that empty tokens in the
// comma-separated list are ignored.
func TestUpdateSitesToBeBlocked_SkipsBlanks(t *testing.T) {
	_, sitesPath := setupTestPaths(t)

	UpdateSitesToBeBlocked("reddit.com,,  ,twitter.com")

	data, err := os.ReadFile(sitesPath)
	if err != nil {
		t.Fatalf("reading sites file: %v", err)
	}

	got := string(data)
	if strings.Contains(got, "127.0.0.1  ") {
		t.Errorf("sites file contains blank entry:\n%s", got)
	}
}

// TestStartBlocking verifies that StartBlocking appends the block markers and
// site entries to the hosts file.
func TestStartBlocking(t *testing.T) {
	hostsPath, sitesPath := setupTestPaths(t)

	initialHosts := "127.0.0.1 localhost\n"
	if err := os.WriteFile(hostsPath, []byte(initialHosts), 0644); err != nil {
		t.Fatalf("writing hosts file: %v", err)
	}

	siteEntries := "127.0.0.1 reddit.com\n127.0.0.1 twitter.com\n"
	if err := os.WriteFile(sitesPath, []byte(siteEntries), 0644); err != nil {
		t.Fatalf("writing sites file: %v", err)
	}

	StartBlocking()

	data, err := os.ReadFile(hostsPath)
	if err != nil {
		t.Fatalf("reading hosts file: %v", err)
	}
	got := string(data)

	if !strings.Contains(got, initialHosts) {
		t.Error("hosts file lost its original content after StartBlocking")
	}
	if !strings.Contains(got, starter) {
		t.Errorf("hosts file missing start marker %q", starter)
	}
	if !strings.Contains(got, ender) {
		t.Errorf("hosts file missing end marker %q", ender)
	}
	if !strings.Contains(got, "127.0.0.1 reddit.com") {
		t.Error("hosts file missing blocked site entry")
	}
}

// TestStopBlocking verifies that StopBlocking removes the block list while
// preserving the rest of the hosts file.
func TestStopBlocking(t *testing.T) {
	hostsPath, _ := setupTestPaths(t)

	hostsContent := "127.0.0.1 localhost\n" +
		"\n" + starter + "\n" +
		"127.0.0.1 reddit.com\n" +
		ender + "\n" +
		"# some trailing comment\n"

	if err := os.WriteFile(hostsPath, []byte(hostsContent), 0644); err != nil {
		t.Fatalf("writing hosts file: %v", err)
	}

	StopBlocking()

	data, err := os.ReadFile(hostsPath)
	if err != nil {
		t.Fatalf("reading hosts file: %v", err)
	}
	got := string(data)

	if strings.Contains(got, starter) {
		t.Error("hosts file still contains start marker after StopBlocking")
	}
	if strings.Contains(got, ender) {
		t.Error("hosts file still contains end marker after StopBlocking")
	}
	if strings.Contains(got, "reddit.com") {
		t.Error("hosts file still contains blocked site after StopBlocking")
	}
	if !strings.Contains(got, "127.0.0.1 localhost") {
		t.Error("hosts file lost original content after StopBlocking")
	}
}

// TestStopBlocking_NoBlockList verifies that StopBlocking is a no-op when
// no block markers exist in the hosts file.
func TestStopBlocking_NoBlockList(t *testing.T) {
	hostsPath, _ := setupTestPaths(t)

	original := "127.0.0.1 localhost\n::1 localhost\n"
	if err := os.WriteFile(hostsPath, []byte(original), 0644); err != nil {
		t.Fatalf("writing hosts file: %v", err)
	}

	StopBlocking()

	data, err := os.ReadFile(hostsPath)
	if err != nil {
		t.Fatalf("reading hosts file: %v", err)
	}
	if string(data) != original {
		t.Errorf("hosts file was modified unexpectedly\ngot: %q\nwant: %q", string(data), original)
	}
}

// TestStartThenStop verifies that StartBlocking followed by StopBlocking
// leaves the hosts file identical to its initial state.
func TestStartThenStop(t *testing.T) {
	hostsPath, sitesPath := setupTestPaths(t)

	original := "127.0.0.1 localhost\n"
	if err := os.WriteFile(hostsPath, []byte(original), 0644); err != nil {
		t.Fatalf("writing hosts file: %v", err)
	}
	if err := os.WriteFile(sitesPath, []byte("127.0.0.1 reddit.com\n"), 0644); err != nil {
		t.Fatalf("writing sites file: %v", err)
	}

	StartBlocking()
	StopBlocking()

	data, err := os.ReadFile(hostsPath)
	if err != nil {
		t.Fatalf("reading hosts file: %v", err)
	}
	// After stop, the file should equal the original (trailing newline may vary by one).
	if !strings.HasPrefix(string(data), original) {
		t.Errorf("hosts file after start+stop:\ngot:  %q\nwant: %q", string(data), original)
	}
	if strings.Contains(string(data), "reddit.com") {
		t.Error("blocked site still present after StopBlocking")
	}
}

// TestProcessMessage_Start verifies that the "start" action calls StartBlocking.
// We detect this by checking whether the hosts file is modified.
func TestProcessMessage_Start(t *testing.T) {
	hostsPath, sitesPath := setupTestPaths(t)

	if err := os.WriteFile(hostsPath, []byte("127.0.0.1 localhost\n"), 0644); err != nil {
		t.Fatalf("writing hosts file: %v", err)
	}
	if err := os.WriteFile(sitesPath, []byte("127.0.0.1 reddit.com\n"), 0644); err != nil {
		t.Fatalf("writing sites file: %v", err)
	}

	ProcessMessage(&ipc.Message{Action: "start", Content: ""})

	data, _ := os.ReadFile(hostsPath)
	if !strings.Contains(string(data), starter) {
		t.Error("ProcessMessage('start') did not trigger StartBlocking")
	}
}

// TestProcessMessage_Stop verifies that the "stop" action calls StopBlocking.
func TestProcessMessage_Stop(t *testing.T) {
	hostsPath, _ := setupTestPaths(t)

	content := "127.0.0.1 localhost\n" + starter + "\n127.0.0.1 reddit.com\n" + ender + "\n"
	if err := os.WriteFile(hostsPath, []byte(content), 0644); err != nil {
		t.Fatalf("writing hosts file: %v", err)
	}

	ProcessMessage(&ipc.Message{Action: "stop", Content: ""})

	data, _ := os.ReadFile(hostsPath)
	if strings.Contains(string(data), starter) {
		t.Error("ProcessMessage('stop') did not trigger StopBlocking")
	}
}

// TestRoundTrip_ExactEquality verifies that the hosts file is byte-for-byte identical
// to its original content after a StartBlocking + StopBlocking cycle.
func TestRoundTrip_ExactEquality(t *testing.T) {
	hostsPath, sitesPath := setupTestPaths(t)

	original := "127.0.0.1 localhost\n::1 localhost\n"
	if err := os.WriteFile(hostsPath, []byte(original), 0644); err != nil {
		t.Fatalf("writing hosts file: %v", err)
	}
	if err := os.WriteFile(sitesPath, []byte("127.0.0.1 reddit.com\n"), 0644); err != nil {
		t.Fatalf("writing sites file: %v", err)
	}

	StartBlocking()
	StopBlocking()

	got, err := os.ReadFile(hostsPath)
	if err != nil {
		t.Fatalf("reading hosts file: %v", err)
	}
	if string(got) != original {
		t.Errorf("hosts file not identical after block+unblock\ngot:  %q\nwant: %q", string(got), original)
	}
}

// TestRoundTrip_NoTrailingNewlines verifies no extra newlines are left behind
// after blocking and unblocking.
func TestRoundTrip_NoTrailingNewlines(t *testing.T) {
	hostsPath, sitesPath := setupTestPaths(t)

	original := "# Static table lookup for hostnames.\n# See hosts(5) for details.\n127.0.0.1        localhost\n::1              localhost\n\n"
	if err := os.WriteFile(hostsPath, []byte(original), 0644); err != nil {
		t.Fatalf("writing hosts file: %v", err)
	}
	if err := os.WriteFile(sitesPath, []byte("127.0.0.1 reddit.com\n127.0.0.1 twitter.com\n"), 0644); err != nil {
		t.Fatalf("writing sites file: %v", err)
	}

	StartBlocking()
	StopBlocking()

	got, err := os.ReadFile(hostsPath)
	if err != nil {
		t.Fatalf("reading hosts file: %v", err)
	}
	
	if string(got) != original {
		t.Errorf("hosts file not identical after block+unblock\ngot:  %q\nwant: %q", string(got), original)
	}
}

// TestRoundTrip_MultipleBlockUnblockCycles verifies that repeated block/unblock
// cycles do not accumulate extra newlines or content.
func TestRoundTrip_MultipleBlockUnblockCycles(t *testing.T) {
	hostsPath, sitesPath := setupTestPaths(t)

	original := "127.0.0.1 localhost\n::1 localhost\n"
	if err := os.WriteFile(hostsPath, []byte(original), 0644); err != nil {
		t.Fatalf("writing hosts file: %v", err)
	}
	if err := os.WriteFile(sitesPath, []byte("127.0.0.1 reddit.com\n"), 0644); err != nil {
		t.Fatalf("writing sites file: %v", err)
	}

	for i := range 3 {
		StartBlocking()
		StopBlocking()

		got, err := os.ReadFile(hostsPath)
		if err != nil {
			t.Fatalf("cycle %d: reading hosts file: %v", i+1, err)
		}
		if string(got) != original {
			t.Errorf("cycle %d: hosts file not identical after block+unblock\ngot:  %q\nwant: %q", i+1, string(got), original)
		}
	}
}

// TestRoundTrip_WithExistingTrailingContent verifies that content after the block
// section is preserved and the file is otherwise identical to the original.
func TestRoundTrip_WithExistingTrailingContent(t *testing.T) {
	hostsPath, sitesPath := setupTestPaths(t)

	original := "127.0.0.1 localhost\n# custom rule\n192.168.1.1 mydevice\n"
	if err := os.WriteFile(hostsPath, []byte(original), 0644); err != nil {
		t.Fatalf("writing hosts file: %v", err)
	}
	if err := os.WriteFile(sitesPath, []byte("127.0.0.1 reddit.com\n"), 0644); err != nil {
		t.Fatalf("writing sites file: %v", err)
	}

	StartBlocking()
	StopBlocking()

	got, err := os.ReadFile(hostsPath)
	if err != nil {
		t.Fatalf("reading hosts file: %v", err)
	}
	if string(got) != original {
		t.Errorf("hosts file not identical after block+unblock\ngot:  %q\nwant: %q", string(got), original)
	}
}

// TestProcessMessage_Update verifies that the "update" action calls UpdateSitesToBeBlocked.
func TestProcessMessage_Update(t *testing.T) {
	_, sitesPath := setupTestPaths(t)

	ProcessMessage(&ipc.Message{Action: "update", Content: "example.com,news.ycombinator.com"})

	data, err := os.ReadFile(sitesPath)
	if err != nil {
		t.Fatalf("reading sites file: %v", err)
	}
	if !strings.Contains(string(data), "127.0.0.1 example.com") {
		t.Error("ProcessMessage('update') did not write sites file correctly")
	}
}

package notarytool

import (
	"bytes"
	"slices"
	"strings"
	"testing"

	"github.com/worldiety/wdyapppack/app/distribution"
	"github.com/worldiety/wdyapppack/foundation/run"
	"github.com/worldiety/wdyapppack/foundation/run/runtest"
)

func TestSubmitAskForTheVerdictAndHideThePassword(t *testing.T) {
	x := runtest.Fake(t, t.TempDir(), "xcrun").Says(t, `{"id":"2efe2717","status":"Accepted","message":"Processing complete"}`)
	var log bytes.Buffer
	n := New(&run.Runner{Log: &log, Verbose: func() bool { return true }}, Tools{Xcrun: x.Path, OS: "darwin"})
	if err := n.Ready(); err != nil {
		t.Fatal(err)
	}
	sub, err := n.Submit("My.zip", distribution.Credentials{AppleID: "dev@example.com", TeamID: "TEAM", Password: "hunter2"})
	if err != nil || sub.ID != "2efe2717" || sub.Status != "Accepted" {
		t.Fatalf("%+v %v", sub, err)
	}
	calls := x.Calls(t)
	want := []string{"notarytool", "submit", "My.zip", "--wait", "--output-format", "json", "--apple-id", "dev@example.com", "--team-id", "TEAM", "--password", "hunter2"}
	if !slices.Equal(calls[1], want) {
		t.Fatalf("%q", calls[1])
	}
	if strings.Contains(log.String(), "hunter2") || !strings.Contains(log.String(), "--password '***'") && !strings.Contains(log.String(), "--password ***") {
		t.Fatalf("the password was shown: %s", log.String())
	}

	if _, err := n.Submit("My.zip", distribution.Credentials{KeychainProfile: "p"}); err != nil {
		t.Fatal(err)
	}
	if got := x.Calls(t)[2]; !slices.Contains(got, "--keychain-profile") {
		t.Fatalf("%q", got)
	}
	if err := n.Staple("My.app"); err != nil {
		t.Fatal(err)
	}
	if got := x.Calls(t)[3]; !slices.Equal(got, []string{"stapler", "staple", "My.app"}) {
		t.Fatalf("%q", got)
	}
}

func TestRejectedVerdictIsStillAVerdict(t *testing.T) {
	x := runtest.Fake(t, t.TempDir(), "xcrun").Says(t, `{"id":"abc","status":"Invalid"}`).Fails(t, 1)
	n := New(&run.Runner{}, Tools{Xcrun: x.Path, OS: "darwin"})
	sub, err := n.Submit("My.zip", distribution.Credentials{KeychainProfile: "p"})
	if err != nil || sub.Status != "Invalid" {
		t.Fatalf("%+v %v", sub, err)
	}
}

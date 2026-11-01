package netrasp

import (
	"context"
	"io"
	"strings"
	"testing"
)

type fakeEnableConn struct {
	sent    []string
	scripts []string
	recvIdx int
}

func (f *fakeEnableConn) Dial(ctx context.Context) error          { return nil }
func (f *fakeEnableConn) Close(ctx context.Context) error         { return nil }
func (f *fakeEnableConn) Send(ctx context.Context, command string) error {
	f.sent = append(f.sent, command)
	return nil
}
func (f *fakeEnableConn) Recv(ctx context.Context) io.Reader {
	s := f.scripts[f.recvIdx]
	f.recvIdx++
	return strings.NewReader(s)
}
func (f *fakeEnableConn) GetHost() *host {
	return &host{enableSecret: "secret123"}
}

// Already privileged (privilege 15): 'enable' returns to the '#' prompt with no
// password prompt, so no secret must be sent.
func TestEnableAlreadyPrivileged(t *testing.T) {
	f := &fakeEnableConn{scripts: []string{
		"spb-kr5-2-01#enable\nspb-kr5-2-01#",
	}}

	if err := enablePrivileges(context.Background(), f, "secret123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.sent) != 1 || f.sent[0] != "enable" {
		t.Fatalf("expected only 'enable' to be sent, got %v", f.sent)
	}
}

// Not privileged: 'enable' triggers a "Password:" prompt, so the secret is sent.
func TestEnableWithPasswordPrompt(t *testing.T) {
	f := &fakeEnableConn{scripts: []string{
		"dcm4500x>enable\nPassword: ",
		"dcm4500x#",
	}}

	if err := enablePrivileges(context.Background(), f, "secret123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.sent) != 2 || f.sent[0] != "enable" || f.sent[1] != "secret123" {
		t.Fatalf("expected 'enable' then secret, got %v", f.sent)
	}
}

// 'enable' with no secret configured but already privileged: nothing to send.
func TestEnableNoSecretAlreadyPrivileged(t *testing.T) {
	f := &fakeEnableConn{scripts: []string{
		"router#enable\nrouter#",
	}}

	if err := enablePrivileges(context.Background(), f, ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.sent) != 1 || f.sent[0] != "enable" {
		t.Fatalf("expected only 'enable' to be sent, got %v", f.sent)
	}
}

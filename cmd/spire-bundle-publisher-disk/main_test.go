package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	configv1 "github.com/spiffe/spire-plugin-sdk/proto/spire/service/common/config/v1"
)

func TestValidateConfiguration(t *testing.T) {
	p := &publisher{}
	valid, err := p.Validate(context.Background(), &configv1.ValidateRequest{HclConfiguration: `directory = "/var/lib/spire/bundles"`})
	if err != nil || !valid.GetValid() {
		t.Fatalf("expected valid configuration, got %#v, %v", valid, err)
	}

	invalid, err := p.Validate(context.Background(), &configv1.ValidateRequest{HclConfiguration: `directory = "/var/lib/spire/bundles"
format = "xml"`})
	if err != nil || invalid.GetValid() {
		t.Fatalf("expected invalid format, got %#v, %v", invalid, err)
	}
}

func TestWriteAtomicCreatesAndReplaces(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "bundle.pem")
	if err := writeAtomic(path, []byte("first"), 0644); err != nil {
		t.Fatal(err)
	}
	if got, err := os.ReadFile(path); err != nil || string(got) != "first" {
		t.Fatalf("first write = %q, %v", got, err)
	}
	if err := writeAtomic(path, []byte("second"), 0644); err != nil {
		t.Fatal(err)
	}
	if got, err := os.ReadFile(path); err != nil || string(got) != "second" {
		t.Fatalf("replacement write = %q, %v", got, err)
	}
}

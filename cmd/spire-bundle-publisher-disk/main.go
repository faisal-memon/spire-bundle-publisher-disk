package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcl"
	"github.com/spiffe/spire-plugin-sdk/pluginmain"
	"github.com/spiffe/spire-plugin-sdk/pluginsdk/support/bundleformat"
	bundlepublisherv1 "github.com/spiffe/spire-plugin-sdk/proto/spire/plugin/server/bundlepublisher/v1"
	"github.com/spiffe/spire-plugin-sdk/proto/spire/plugin/types"
	configv1 "github.com/spiffe/spire-plugin-sdk/proto/spire/service/common/config/v1"
)

type config struct {
	Directory string `hcl:"directory"`
	Filename  string `hcl:"filename"`
	Format    string `hcl:"format"`
}

type publisher struct {
	bundlepublisherv1.UnsafeBundlePublisherServer
	configv1.UnsafeConfigServer
	log hclog.Logger
	cfg config
}

func (p *publisher) SetLogger(log hclog.Logger) { p.log = log }

func (p *publisher) Validate(_ context.Context, req *configv1.ValidateRequest) (*configv1.ValidateResponse, error) {
	cfg := config{Format: "pem", Filename: "bundle.pem"}
	if err := hcl.Decode(&cfg, req.GetHclConfiguration()); err != nil {
		return &configv1.ValidateResponse{Valid: false, Notes: []string{err.Error()}}, nil
	}
	if cfg.Directory == "" {
		return &configv1.ValidateResponse{Valid: false, Notes: []string{"directory is required"}}, nil
	}
	if cfg.Format != "pem" && cfg.Format != "spiffe" && cfg.Format != "jwks" {
		return &configv1.ValidateResponse{Valid: false, Notes: []string{"unsupported format: " + cfg.Format}}, nil
	}
	if filepath.Base(cfg.Filename) != cfg.Filename {
		return &configv1.ValidateResponse{Valid: false, Notes: []string{"filename must not contain a directory"}}, nil
	}
	return &configv1.ValidateResponse{Valid: true}, nil
}

func (p *publisher) Configure(_ context.Context, req *configv1.ConfigureRequest) (*configv1.ConfigureResponse, error) {
	cfg := config{Format: "pem", Filename: "bundle.pem"}
	if err := hcl.Decode(&cfg, req.GetHclConfiguration()); err != nil {
		return nil, fmt.Errorf("decode configuration: %w", err)
	}
	if cfg.Directory == "" {
		return nil, fmt.Errorf("directory is required")
	}
	if cfg.Format != "pem" && cfg.Format != "spiffe" && cfg.Format != "jwks" {
		return nil, fmt.Errorf("unsupported format %q", cfg.Format)
	}
	if filepath.Base(cfg.Filename) != cfg.Filename {
		return nil, fmt.Errorf("filename must not contain a directory")
	}
	p.cfg = cfg
	return &configv1.ConfigureResponse{}, nil
}

func (p *publisher) PublishBundle(_ context.Context, req *bundlepublisherv1.PublishBundleRequest) (*bundlepublisherv1.PublishBundleResponse, error) {
	if req.GetBundle() == nil {
		return nil, fmt.Errorf("bundle is required")
	}
	format, err := bundleformat.FromString(p.cfg.Format)
	if err != nil {
		return nil, err
	}
	data, err := bundleformat.FormatBundle(req.GetBundle(), format)
	if err != nil {
		return nil, fmt.Errorf("format bundle: %w", err)
	}
	if err := writeAtomic(filepath.Join(p.cfg.Directory, p.cfg.Filename), data, 0644); err != nil {
		return nil, err
	}
	return &bundlepublisherv1.PublishBundleResponse{}, nil
}

func writeAtomic(path string, data []byte, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".bundle-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(mode); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

func main() {
	p := &publisher{}
	pluginmain.Serve(bundlepublisherv1.BundlePublisherPluginServer(p), configv1.ConfigServiceServer(p))
}

var _ = types.Bundle{}

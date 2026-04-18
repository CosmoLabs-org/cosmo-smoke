package goss

import (
	"os"
	"testing"
)

func TestParseBasic(t *testing.T) {
	data, err := os.ReadFile("testdata/goss/basic.yaml")
	if err != nil {
		t.Fatalf("reading test fixture: %v", err)
	}

	gf, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(gf.Process) != 2 {
		t.Errorf("Process count = %d, want 2", len(gf.Process))
	}
	if !boolVal(gf.Process["nginx"], "running") {
		t.Error("nginx.running should be true")
	}

	if len(gf.Port) != 3 {
		t.Errorf("Port count = %d, want 3", len(gf.Port))
	}

	if len(gf.Command) != 1 {
		t.Errorf("Command count = %d, want 1", len(gf.Command))
	}

	if len(gf.File) != 2 {
		t.Errorf("File count = %d, want 2", len(gf.File))
	}

	if len(gf.HTTP) != 1 {
		t.Errorf("HTTP count = %d, want 1", len(gf.HTTP))
	}

	if len(gf.Package) != 2 {
		t.Errorf("Package count = %d, want 2", len(gf.Package))
	}

	if len(gf.Service) != 1 {
		t.Errorf("Service count = %d, want 1", len(gf.Service))
	}
}

func TestParseLongtail(t *testing.T) {
	data, err := os.ReadFile("testdata/goss/longtail.yaml")
	if err != nil {
		t.Fatalf("reading test fixture: %v", err)
	}

	gf, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(gf.User) != 1 {
		t.Errorf("User count = %d, want 1", len(gf.User))
	}
	if len(gf.Group) != 1 {
		t.Errorf("Group count = %d, want 1", len(gf.Group))
	}
	if len(gf.DNS) != 1 {
		t.Errorf("DNS count = %d, want 1", len(gf.DNS))
	}
	if len(gf.Addr) != 1 {
		t.Errorf("Addr count = %d, want 1", len(gf.Addr))
	}
	if len(gf.Interface) != 1 {
		t.Errorf("Interface count = %d, want 1", len(gf.Interface))
	}
	if len(gf.Mount) != 1 {
		t.Errorf("Mount count = %d, want 1", len(gf.Mount))
	}
	if len(gf.KernelParam) != 1 {
		t.Errorf("KernelParam count = %d, want 1", len(gf.KernelParam))
	}
}

func TestParseInvalid(t *testing.T) {
	_, err := Parse([]byte(": invalid yaml : ["))
	if err == nil {
		t.Error("Parse() should fail on invalid YAML")
	}
}

func TestParseEmpty(t *testing.T) {
	gf, err := Parse([]byte(""))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(gf.Process) != 0 {
		t.Error("Empty YAML should produce empty GossFile")
	}
}

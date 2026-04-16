package detector

import (
	"testing"
)

func TestParseSSOutput(t *testing.T) {
	output := `State   Recv-Q  Send-Q  Local Address:Port   Peer Address:Port
LISTEN  0       128           *:8080              *:*
LISTEN  0       128           *:3000              *:*
LISTEN  0       128     [::]:22                [::]:*`

	ports, err := parseSSOutput(output)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(ports) != 3 {
		t.Errorf("expected 3 ports, got %d", len(ports))
	}

	// Check first port
	found8080 := false
	for _, p := range ports {
		if p.Port == 8080 {
			found8080 = true
			if p.Protocol != "tcp" {
				t.Errorf("expected tcp protocol, got %s", p.Protocol)
			}
		}
	}
	if !found8080 {
		t.Error("expected to find port 8080")
	}
}

func TestParseNetstatOutput(t *testing.T) {
	output := `Active Internet connections (only servers)
Proto Recv-Q Send-Q Local Address           Foreign Address         State       PID/Program name
tcp        0      0 0.0.0.0:8080            0.0.0.0:*               LISTEN      1/node
tcp        0      0 0.0.0.0:443             0.0.0.0:*               LISTEN      1/nginx`

	ports, err := parseNetstatOutput(output)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(ports) != 2 {
		t.Errorf("expected 2 ports, got %d", len(ports))
	}
}

func TestParseDockerPortOutput(t *testing.T) {
	output := `8080/tcp -> 0.0.0.0:8080
443/tcp -> 0.0.0.0:443
53/udp -> 0.0.0.0:53`

	ports, err := parseDockerPortOutput(output)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(ports) != 3 {
		t.Errorf("expected 3 ports, got %d", len(ports))
	}

	// Check UDP port
	foundUDP := false
	for _, p := range ports {
		if p.Port == 53 && p.Protocol == "udp" {
			foundUDP = true
		}
	}
	if !foundUDP {
		t.Error("expected to find UDP port 53")
	}
}

func TestIsHTTPPort(t *testing.T) {
	tests := []struct {
		port     int
		expected bool
	}{
		{80, true},
		{443, true},
		{8080, true},
		{3000, true},
		{22, false},
		{5432, false},
		{27017, false},
	}

	for _, tc := range tests {
		if got := isHTTPPort(tc.port); got != tc.expected {
			t.Errorf("isHTTPPort(%d) = %v, want %v", tc.port, got, tc.expected)
		}
	}
}

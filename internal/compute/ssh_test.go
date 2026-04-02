package compute

import (
	"testing"
)

func TestSSHArgs(t *testing.T) {
	tests := []struct {
		name      string
		user      string
		ip        string
		extra     []string
		wantLast  string
		wantCount int
	}{
		{
			name:      "basic",
			user:      "ali",
			ip:        "35.1.2.3",
			wantLast:  "ali@35.1.2.3",
			wantCount: 5,
		},
		{
			name:      "no user",
			ip:        "35.1.2.3",
			wantLast:  "35.1.2.3",
			wantCount: 5,
		},
		{
			name:      "with extra args",
			user:      "root",
			ip:        "10.0.0.1",
			extra:     []string{"-L", "8080:localhost:80"},
			wantLast:  "8080:localhost:80",
			wantCount: 7, // 4 options + target + 2 extra
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := SSHArgs(tt.user, tt.ip, tt.extra)
			if len(args) != tt.wantCount {
				t.Errorf("arg count: got %d, want %d: %v", len(args), tt.wantCount, args)
			}
			if args[len(args)-1] != tt.wantLast {
				t.Errorf("last arg: got %q, want %q", args[len(args)-1], tt.wantLast)
			}
			// Verify no shell injection vectors
			for _, a := range args {
				if a == "" {
					t.Error("empty arg detected — potential injection")
				}
			}
		})
	}
}

func TestSCPArgs(t *testing.T) {
	tests := []struct {
		name    string
		user    string
		ip      string
		src     string
		dst     string
		wantSrc string
		wantDst string
	}{
		{
			name:    "local to remote",
			user:    "ali",
			ip:      "35.1.2.3",
			src:     "/tmp/file.txt",
			dst:     "myvm:/home/ali/file.txt",
			wantSrc: "/tmp/file.txt",
			wantDst: "ali@35.1.2.3:/home/ali/file.txt",
		},
		{
			name:    "remote to local",
			user:    "",
			ip:      "10.0.0.1",
			src:     "myvm:/etc/hosts",
			dst:     "/tmp/hosts",
			wantSrc: "10.0.0.1:/etc/hosts",
			wantDst: "/tmp/hosts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := SCPArgs(tt.user, tt.ip, tt.src, tt.dst)
			// Last two args are src and dst
			gotSrc := args[len(args)-2]
			gotDst := args[len(args)-1]
			if gotSrc != tt.wantSrc {
				t.Errorf("src: got %q, want %q", gotSrc, tt.wantSrc)
			}
			if gotDst != tt.wantDst {
				t.Errorf("dst: got %q, want %q", gotDst, tt.wantDst)
			}
		})
	}
}

func TestResolveInstanceIP(t *testing.T) {
	tests := []struct {
		name    string
		inst    *Instance
		wantIP  string
		wantErr bool
	}{
		{
			name:   "prefer external",
			inst:   &Instance{Name: "vm", ExternalIP: "35.1.2.3", InternalIP: "10.0.0.1"},
			wantIP: "35.1.2.3",
		},
		{
			name:   "fallback to internal",
			inst:   &Instance{Name: "vm", InternalIP: "10.0.0.1"},
			wantIP: "10.0.0.1",
		},
		{
			name:    "no IP",
			inst:    &Instance{Name: "vm"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockClient{
				instanceMap: map[string]*Instance{
					"vm": tt.inst,
				},
			}
			ip, err := ResolveInstanceIP(t.Context(), mock, "proj", "zone", "vm")
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ip != tt.wantIP {
				t.Errorf("ip: got %q, want %q", ip, tt.wantIP)
			}
		})
	}
}

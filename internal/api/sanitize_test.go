package api

import (
	"testing"
)

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "normal string",
			input: "test string",
			want:  "test string",
		},
		{
			name:  "with whitespace",
			input: "  test  ",
			want:  "test",
		},
		{
			name:  "with HTML entities",
			input: "&lt;script&gt;alert('xss')&lt;/script&gt;",
			want:  "<script>alert('xss')</script>",
		},
		{
			name:  "with null bytes",
			input: "test\x00string",
			want:  "teststring",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "only whitespace",
			input: "   ",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeString(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSanitizeURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		want    string
	}{
		{
			name:    "valid http URL",
			input:   "http://example.com",
			wantErr: false,
			want:    "http://example.com",
		},
		{
			name:    "valid https URL",
			input:   "https://example.com/path",
			wantErr: false,
			want:    "https://example.com/path",
		},
		{
			name:    "invalid scheme",
			input:   "ftp://example.com",
			wantErr: true,
		},
		{
			name:    "invalid URL",
			input:   "not a url",
			wantErr: true,
		},
		{
			name:    "with whitespace",
			input:   "  https://example.com  ",
			wantErr: false,
			want:    "https://example.com",
		},
		{
			name:    "with query params",
			input:   "https://example.com?param=value",
			wantErr: false,
			want:    "https://example.com?param=value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SanitizeURL(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("SanitizeURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("SanitizeURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSanitizeHostname(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "normal hostname",
			input: "example.com",
			want:  "example.com",
		},
		{
			name:  "with http prefix",
			input: "http://example.com",
			want:  "example.com",
		},
		{
			name:  "with https prefix",
			input: "https://example.com",
			want:  "example.com",
		},
		{
			name:  "with trailing slash",
			input: "example.com/",
			want:  "example.com",
		},
		{
			name:  "with path",
			input: "example.com/path/to/resource",
			want:  "example.com",
		},
		{
			name:  "with uppercase",
			input: "EXAMPLE.COM",
			want:  "example.com",
		},
		{
			name:  "with whitespace",
			input: "  example.com  ",
			want:  "example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeHostname(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeHostname() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSanitizeDomain(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "normal domain",
			input: "example.com",
			want:  "example.com",
		},
		{
			name:  "with port",
			input: "example.com:8080",
			want:  "example.com",
		},
		{
			name:  "with protocol and port",
			input: "https://example.com:8080/path",
			want:  "example.com",
		},
		{
			name:  "with uppercase",
			input: "EXAMPLE.COM",
			want:  "example.com",
		},
		{
			name:  "subdomain",
			input: "api.example.com",
			want:  "api.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeDomain(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeDomain() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "normal filename",
			input: "file.txt",
			want:  "file.txt",
		},
		{
			name:  "with path traversal",
			input: "../../etc/passwd",
			want:  "etcpasswd",
		},
		{
			name:  "with backslash",
			input: "..\\..\\file.txt",
			want:  "file.txt",
		},
		{
			name:  "with null bytes",
			input: "file\x00.txt",
			want:  "file.txt",
		},
		{
			name:  "with whitespace",
			input: "  file.txt  ",
			want:  "file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeFilename(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeFilename() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSanitizeEnvironmentVariableKey(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "normal key",
			input: "DATABASE_URL",
			want:  "DATABASE_URL",
		},
		{
			name:  "lowercase",
			input: "database_url",
			want:  "DATABASE_URL",
		},
		{
			name:  "with special chars",
			input: "DATABASE-URL!@#",
			want:  "DATABASEURL",
		},
		{
			name:  "with whitespace",
			input: "  database url  ",
			want:  "DATABASEURL",
		},
		{
			name:  "with numbers",
			input: "API_KEY_123",
			want:  "API_KEY_123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeEnvironmentVariableKey(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeEnvironmentVariableKey() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSanitizeGitBranch(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "normal branch",
			input: "main",
			want:  "main",
		},
		{
			name:  "with slashes",
			input: "feature/new-feature",
			want:  "feature/new-feature",
		},
		{
			name:  "with dangerous chars",
			input: "branch<script>",
			want:  "branchscript",
		},
		{
			name:  "leading dots",
			input: "../branch",
			want:  "branch",
		},
		{
			name:  "trailing slashes",
			input: "branch/",
			want:  "branch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeGitBranch(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeGitBranch() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSanitizeCommitSHA(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "valid SHA-1",
			input: "abc123def456",
			want:  "abc123def456",
		},
		{
			name:  "with uppercase",
			input: "ABC123DEF456",
			want:  "abc123def456",
		},
		{
			name:  "with invalid chars",
			input: "abc123-xyz",
			want:  "abc123", // "-" and "xyz" are removed (not hex)
		},
		{
			name:  "too long",
			input: string(make([]byte, 100)),
			want:  "", // Only hex chars, so empty after filtering
		},
		{
			name:  "with whitespace",
			input: "  abc123  ",
			want:  "abc123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeCommitSHA(tt.input)
			// For the "too long" case, we can't predict the exact output
			// but it should be <= 64 chars and only contain hex chars
			if tt.name == "too long" {
				if len(got) > 64 {
					t.Errorf("SanitizeCommitSHA() length = %d, want <= 64", len(got))
				}
				// Verify it only contains hex chars
				for _, r := range got {
					if !((r >= 'a' && r <= 'f') || (r >= '0' && r <= '9')) {
						t.Errorf("SanitizeCommitSHA() contains non-hex char: %c", r)
					}
				}
			} else {
				// For other cases, check the expected output
				if got != tt.want {
					t.Errorf("SanitizeCommitSHA() = %q, want %q", got, tt.want)
				}
			}
		})
	}
}


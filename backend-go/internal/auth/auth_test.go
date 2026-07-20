package auth

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPassword(t *testing.T) {
	hash, err := HashPassword("testpassword")
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if !Password("testpassword", hash) {
		t.Fatal("expected password to match hash")
	}

	if Password("wrongpassword", hash) {
		t.Fatal("expected wrong password to not match hash")
	}
}

func TestParseMethod(t *testing.T) {
	tests := []struct {
		input string
		want  Method
	}{
		{"password", MethodPassword},
		{"PASSWORD", MethodPassword},
		{"htpasswd", MethodHtpasswd},
		{"HTPASSWD", MethodHtpasswd},
		{"ldap", MethodLDAP},
		{"LDAP", MethodLDAP},
		{"unknown", MethodPassword},
		{"", MethodPassword},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := ParseMethod(tt.input); got != tt.want {
				t.Fatalf("ParseMethod(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestHtpasswd(t *testing.T) {
	// Create a temporary htpasswd file.
	dir := t.TempDir()
	htpasswdPath := filepath.Join(dir, ".htpasswd")

	// Generate a bcrypt hash for "testpassword".
	hash, err := HashPassword("testpassword")
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	// Write the htpasswd file.
	content := "testuser:" + hash + "\n"
	if err := os.WriteFile(htpasswdPath, []byte(content), 0644); err != nil {
		t.Fatalf("writing htpasswd file: %v", err)
	}

	// Test valid credentials.
	ok, err := Htpasswd("testuser", "testpassword", htpasswdPath)
	if err != nil {
		t.Fatalf("Htpasswd failed: %v", err)
	}
	if !ok {
		t.Fatal("expected valid credentials to authenticate")
	}

	// Test invalid password.
	ok, err = Htpasswd("testuser", "wrongpassword", htpasswdPath)
	if err != nil {
		t.Fatalf("Htpasswd failed: %v", err)
	}
	if ok {
		t.Fatal("expected invalid password to not authenticate")
	}

	// Test non-existent user.
	ok, err = Htpasswd("unknown", "testpassword", htpasswdPath)
	if err != nil {
		t.Fatalf("Htpasswd failed: %v", err)
	}
	if ok {
		t.Fatal("expected non-existent user to not authenticate")
	}

	// Test non-existent file.
	_, err = Htpasswd("testuser", "testpassword", "/nonexistent/file")
	if err == nil {
		t.Fatal("expected error for non-existent file")
	}
}

func TestSHA256Hash(t *testing.T) {
	hash1 := SHA256Hash("hello")
	hash2 := SHA256Hash("hello")
	hash3 := SHA256Hash("world")

	if hash1 != hash2 {
		t.Fatal("same input should produce same hash")
	}
	if hash1 == hash3 {
		t.Fatal("different input should produce different hash")
	}
	if len(hash1) != 64 {
		t.Fatalf("expected SHA-256 hex string length 64, got %d", len(hash1))
	}
}

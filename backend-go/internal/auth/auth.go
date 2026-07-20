// Package auth handles user authentication (password, htpasswd, ldap).
package auth

import (
	"bufio"
	"crypto/sha256"
	"crypto/tls"
	"fmt"
	"os"
	"strings"

	ldap "github.com/go-ldap/ldap/v3"
	"golang.org/x/crypto/bcrypt"
)

// Method represents an authentication method.
type Method int

const (
	MethodPassword Method = iota
	MethodHtpasswd
	MethodLDAP
)

// ParseMethod parses an authentication method string.
func ParseMethod(s string) Method {
	switch strings.ToLower(s) {
	case "password":
		return MethodPassword
	case "htpasswd":
		return MethodHtpasswd
	case "ldap":
		return MethodLDAP
	default:
		return MethodPassword
	}
}

// Password verifies a password against a bcrypt hash.
func Password(plaintext, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plaintext))
	return err == nil
}

// HashPassword generates a bcrypt hash of the given password.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Htpasswd verifies a username/password pair against an htpasswd file.
func Htpasswd(username, password, path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, fmt.Errorf("opening htpasswd file: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		if parts[0] == username {
			err := bcrypt.CompareHashAndPassword([]byte(parts[1]), []byte(password))
			if err == nil {
				return true, nil
			}
			return false, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("reading htpasswd file: %w", err)
	}

	return false, nil
}

// LDAPAuth authenticates a user against an LDAP server.
func LDAPAuth(username, password, uri string, useStartTLS bool, baseDN, bindDN, bindPass, userFilter string) error {
	// Connect to LDAP server. The v3.4.x library auto-sets LDAP v3 protocol.
	conn, err := ldap.DialURL(uri)
	if err != nil {
		return fmt.Errorf("connecting to LDAP: %w", err)
	}
	defer conn.Close()

	if useStartTLS {
		startTLSConfig := &tls.Config{
			InsecureSkipVerify: false,
		}
		if err := conn.StartTLS(startTLSConfig); err != nil {
			return fmt.Errorf("StartTLS: %w", err)
		}
	}

	// Bind as admin user.
	if err := conn.Bind(bindDN, bindPass); err != nil {
		return fmt.Errorf("admin bind: %w", err)
	}

	// Search for the user.
	filter := strings.Replace(userFilter, "%s", ldap.EscapeFilter(username), 1)
	searchRequest := ldap.NewSearchRequest(
		baseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		filter,
		[]string{"dn"},
		nil,
	)

	searchResult, err := conn.Search(searchRequest)
	if err != nil {
		return fmt.Errorf("LDAP search: %w", err)
	}

	if len(searchResult.Entries) == 0 {
		return fmt.Errorf("user not found or unauthorized")
	}

	if len(searchResult.Entries) > 1 {
		return fmt.Errorf("multiple users matched the filter - LDAP filter too broad")
	}

	userDN := searchResult.Entries[0].DN

	// Bind as the user.
	if err := conn.Bind(userDN, password); err != nil {
		return fmt.Errorf("user authentication failed")
	}

	return nil
}

// SHA256Hash computes a SHA-256 hash of the input string.
func SHA256Hash(input string) string {
	h := sha256.Sum256([]byte(input))
	return fmt.Sprintf("%x", h)
}

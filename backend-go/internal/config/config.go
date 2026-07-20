// Package config handles loading and managing the Hauk server configuration.
// Configuration is loaded from /etc/hauk/config.go or environment variables,
// with env vars taking precedence.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	// BackendVersion is the current backend version, matching the PHP backend.
	BackendVersion = "1.6.2"

	// ConfigFileName is the config file name for parsing.
	ConfigFileName = ""
)

var (
	// ConfigPaths are the paths where config files are searched for.
	ConfigPaths = []string{"/etc/hauk/config.yaml", "/etc/hauk/config.yml", "/etc/hauk/config.json"}

	// SupportedLanguages lists the languages supported by the backend.
	SupportedLanguages = []string{"ca", "de", "en", "eu", "fr", "it", "nb_NO", "nl", "nn", "ro", "ru", "tr", "uk"}
)

// Config holds all configuration values for the Hauk server.
type Config struct {
	// StorageBackend identifies the storage backend ("redis").
	StorageBackend string `yaml:"storage_backend" json:"storage_backend" mapstructure:"storage_backend"`

	// RedisHost is the Redis host or socket path.
	RedisHost string `yaml:"redis_host" json:"redis_host" mapstructure:"redis_host"`

	// RedisPort is the Redis port.
	RedisPort int `yaml:"redis_port" json:"redis_port" mapstructure:"redis_port"`

	// RedisUseAuth indicates whether Redis requires authentication.
	RedisUseAuth bool `yaml:"redis_use_auth" json:"redis_use_auth" mapstructure:"redis_use_auth"`

	// RedisAuth is the Redis password.
	RedisAuth string `yaml:"redis_auth" json:"redis_auth" mapstructure:"redis_auth"`

	// RedisPrefix is prepended to all Redis keys.
	RedisPrefix string `yaml:"redis_prefix" json:"redis_prefix" mapstructure:"redis_prefix"`

	// AuthMethod selects the authentication method ("password", "htpasswd", "ldap").
	AuthMethod string `yaml:"auth_method" json:"auth_method" mapstructure:"auth_method"`

	// PasswordHash is the bcrypt hash for PASSWORD auth.
	PasswordHash string `yaml:"password_hash" json:"password_hash" mapstructure:"password_hash"`

	// HtpasswdPath is the path to the htpasswd file for HTPASSWD auth.
	HtpasswdPath string `yaml:"htpasswd_path" json:"htpasswd_path" mapstructure:"htpasswd_path"`

	// LdapURI is the LDAP server URI.
	LdapURI string `yaml:"ldap_uri" json:"ldap_uri" mapstructure:"ldap_uri"`

	// LdapStartTLS enables StartTLS for LDAP.
	LdapStartTLS bool `yaml:"ldap_start_tls" json:"ldap_start_tls" mapstructure:"ldap_start_tls"`

	// LdapBaseDN is the base DN for LDAP searches.
	LdapBaseDN string `yaml:"ldap_base_dn" json:"ldap_base_dn" mapstructure:"ldap_base_dn"`

	// LdapBindDN is the DN to bind as for user searches.
	LdapBindDN string `yaml:"ldap_bind_dn" json:"ldap_bind_dn" mapstructure:"ldap_bind_dn"`

	// LdapBindPass is the password for the LDAP bind DN.
	LdapBindPass string `yaml:"ldap_bind_pass" json:"ldap_bind_pass" mapstructure:"ldap_bind_pass"`

	// LdapUserFilter is the LDAP filter for finding users (%s is username).
	LdapUserFilter string `yaml:"ldap_user_filter" json:"ldap_user_filter" mapstructure:"ldap_user_filter"`

	// AllowLinkReq allows clients to request custom link IDs.
	AllowLinkReq bool `yaml:"allow_link_req" json:"allow_link_req" mapstructure:"allow_link_req"`

	// ReservedLinks reserves specific link IDs for specific users.
	ReservedLinks map[string][]string `yaml:"reserved_links" json:"reserved_links" mapstructure:"reserved_links"`

	// ReserveWhitelist enforces that only reserved links can be custom.
	ReserveWhitelist bool `yaml:"reserve_whitelist" json:"reserve_whitelist" mapstructure:"reserve_whitelist"`

	// LinkStyle selects the link generation style (0-11).
	LinkStyle int `yaml:"link_style" json:"link_style" mapstructure:"link_style"`

	// MapTileURI is the Leaflet tile URI template.
	MapTileURI string `yaml:"map_tile_uri" json:"map_tile_uri" mapstructure:"map_tile_uri"`

	// MapAttribution is the map attribution HTML.
	MapAttribution string `yaml:"map_attribution" json:"map_attribution" mapstructure:"map_attribution"`

	// DefaultZoom is the default map zoom level (0-20).
	DefaultZoom int `yaml:"default_zoom" json:"default_zoom" mapstructure:"default_zoom"`

	// MaxZoom is the maximum map zoom level.
	MaxZoom int `yaml:"max_zoom" json:"max_zoom" mapstructure:"max_zoom"`

	// MaxDuration is the maximum share duration in seconds.
	MaxDuration int `yaml:"max_duration" json:"max_duration" mapstructure:"max_duration"`

	// MinInterval is the minimum location update interval in seconds.
	MinInterval int `yaml:"min_interval" json:"min_interval" mapstructure:"min_interval"`

	// OfflineTimeout is seconds without updates before marking offline.
	OfflineTimeout int `yaml:"offline_timeout" json:"offline_timeout" mapstructure:"offline_timeout"`

	// RequestTimeout is the HTTP request timeout for the frontend.
	RequestTimeout int `yaml:"request_timeout" json:"request_timeout" mapstructure:"request_timeout"`

	// MaxCachedPts is the maximum number of location points stored per share.
	MaxCachedPts int `yaml:"max_cached_pts" json:"max_cached_pts" mapstructure:"max_cached_pts"`

	// MaxShownPts is the maximum number of points visible on the map.
	MaxShownPts int `yaml:"max_shown_pts" json:"max_shown_pts" mapstructure:"max_shown_pts"`

	// VDataPoints is the number of data points used for velocity calculation.
	VDataPoints int `yaml:"v_data_points" json:"v_data_points" mapstructure:"v_data_points"`

	// TrailColor is the HTML color for map trails.
	TrailColor string `yaml:"trail_color" json:"trail_color" mapstructure:"trail_color"`

	// VelocityUnit selects the velocity unit ("kmh", "mph", "mps").
	VelocityUnit string `yaml:"velocity_unit" json:"velocity_unit" mapstructure:"velocity_unit"`

	// PublicURL is the public base URL of the Hauk instance (with trailing slash).
	PublicURL string `yaml:"public_url" json:"public_url" mapstructure:"public_url"`

	// ServerPort is the port the HTTP server listens on.
	ServerPort int `yaml:"server_port" json:"server_port" mapstructure:"server_port"`
}

// DefaultConfig returns a Config with all defaults populated.
func DefaultConfig() *Config {
	return &Config{
		StorageBackend: "redis",
		RedisHost:      "localhost",
		RedisPort:      6379,
		RedisUseAuth:   false,
		RedisAuth:      "",
		RedisPrefix:    "hauk",
		AuthMethod:     "password",
		PasswordHash:   "$2y$10$4ZP1iY8A3dZygXoPgsXYV.S3gHzBbiT9nSfONjhWrvMxVPkcFq1Ka",
		HtpasswdPath:   "/etc/hauk/users.htpasswd",
		LdapURI:        "ldaps://ldap.example.com:636",
		LdapStartTLS:   false,
		LdapBaseDN:     "ou=People,dc=example,dc=com",
		LdapBindDN:     "cn=admin,dc=example,dc=com",
		LdapBindPass:   "Adm1nP4ssw0rd",
		LdapUserFilter: "(uid=%s)",
		AllowLinkReq:   true,
		ReservedLinks:  make(map[string][]string),
		ReserveWhitelist: false,
		LinkStyle:      0,
		MapTileURI:     "https://tile.openstreetmap.org/{z}/{x}/{y}.png",
		MapAttribution: `Map data &copy; <a href="https://www.openstreetmap.org/">OpenStreetMap</a> contributors, <a href="https://creativecommons.org/licenses/by-sa/2.0/">CC-BY-SA</a>`,
		DefaultZoom:    14,
		MaxZoom:        19,
		MaxDuration:    86400,
		MinInterval:    1,
		OfflineTimeout: 30,
		RequestTimeout: 10,
		MaxCachedPts:   3,
		MaxShownPts:    100,
		VDataPoints:    2,
		TrailColor:     "#d80037",
		VelocityUnit:   "kmh",
		PublicURL:      "https://example.com/",
		ServerPort:     8080,
	}
}

// LoadConfig loads configuration from file and environment variables.
// Environment variables override file configuration values.
func LoadConfig() (*Config, error) {
	cfg := DefaultConfig()

	// Try to load from file.
	for _, path := range ConfigPaths {
		if data, err := os.ReadFile(path); err == nil {
			if err := parseConfig(data, cfg); err != nil {
				return nil, fmt.Errorf("parsing config file %s: %w", path, err)
			}
			break
		}
	}

	// Override with environment variables.
	if v := os.Getenv("HAUK_STORAGE_BACKEND"); v != "" {
		cfg.StorageBackend = v
	}
	if v := os.Getenv("HAUK_REDIS_HOST"); v != "" {
		cfg.RedisHost = v
	}
	if v := os.Getenv("HAUK_REDIS_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.RedisPort = port
		}
	}
	if v := os.Getenv("HAUK_REDIS_USE_AUTH"); v != "" {
		cfg.RedisUseAuth = v == "true" || v == "1"
	}
	if v := os.Getenv("HAUK_REDIS_AUTH"); v != "" {
		cfg.RedisAuth = v
	}
	if v := os.Getenv("HAUK_REDIS_PREFIX"); v != "" {
		cfg.RedisPrefix = v
	}
	if v := os.Getenv("HAUK_AUTH_METHOD"); v != "" {
		cfg.AuthMethod = v
	}
	if v := os.Getenv("HAUK_PASSWORD_HASH"); v != "" {
		cfg.PasswordHash = v
	}
	if v := os.Getenv("HAUK_HTPASSWD_PATH"); v != "" {
		cfg.HtpasswdPath = v
	}
	if v := os.Getenv("HAUK_LDAP_URI"); v != "" {
		cfg.LdapURI = v
	}
	if v := os.Getenv("HAUK_LDAP_START_TLS"); v != "" {
		cfg.LdapStartTLS = v == "true" || v == "1"
	}
	if v := os.Getenv("HAUK_LDAP_BASE_DN"); v != "" {
		cfg.LdapBaseDN = v
	}
	if v := os.Getenv("HAUK_LDAP_BIND_DN"); v != "" {
		cfg.LdapBindDN = v
	}
	if v := os.Getenv("HAUK_LDAP_BIND_PASS"); v != "" {
		cfg.LdapBindPass = v
	}
	if v := os.Getenv("HAUK_LDAP_USER_FILTER"); v != "" {
		cfg.LdapUserFilter = v
	}
	if v := os.Getenv("HAUK_ALLOW_LINK_REQ"); v != "" {
		cfg.AllowLinkReq = v == "true" || v == "1"
	}
	if v := os.Getenv("HAUK_RESERVE_WHITELIST"); v != "" {
		cfg.ReserveWhitelist = v == "true" || v == "1"
	}
	if v := os.Getenv("HAUK_LINK_STYLE"); v != "" {
		if style, err := strconv.Atoi(v); err == nil {
			cfg.LinkStyle = style
		}
	}
	if v := os.Getenv("HAUK_MAP_TILE_URI"); v != "" {
		cfg.MapTileURI = v
	}
	if v := os.Getenv("HAUK_MAP_ATTRIBUTION"); v != "" {
		cfg.MapAttribution = v
	}
	if v := os.Getenv("HAUK_DEFAULT_ZOOM"); v != "" {
		if z, err := strconv.Atoi(v); err == nil {
			cfg.DefaultZoom = z
		}
	}
	if v := os.Getenv("HAUK_MAX_ZOOM"); v != "" {
		if z, err := strconv.Atoi(v); err == nil {
			cfg.MaxZoom = z
		}
	}
	if v := os.Getenv("HAUK_MAX_DURATION"); v != "" {
		if d, err := strconv.Atoi(v); err == nil {
			cfg.MaxDuration = d
		}
	}
	if v := os.Getenv("HAUK_MIN_INTERVAL"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.MinInterval = i
		}
	}
	if v := os.Getenv("HAUK_OFFLINE_TIMEOUT"); v != "" {
		if t, err := strconv.Atoi(v); err == nil {
			cfg.OfflineTimeout = t
		}
	}
	if v := os.Getenv("HAUK_REQUEST_TIMEOUT"); v != "" {
		if t, err := strconv.Atoi(v); err == nil {
			cfg.RequestTimeout = t
		}
	}
	if v := os.Getenv("HAUK_MAX_CACHED_PTS"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			cfg.MaxCachedPts = p
		}
	}
	if v := os.Getenv("HAUK_MAX_SHOWN_PTS"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			cfg.MaxShownPts = p
		}
	}
	if v := os.Getenv("HAUK_V_DATA_POINTS"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			cfg.VDataPoints = p
		}
	}
	if v := os.Getenv("HAUK_TRAIL_COLOR"); v != "" {
		cfg.TrailColor = v
	}
	if v := os.Getenv("HAUK_VELOCITY_UNIT"); v != "" {
		cfg.VelocityUnit = v
	}
	if v := os.Getenv("HAUK_PUBLIC_URL"); v != "" {
		cfg.PublicURL = v
	}
	if v := os.Getenv("HAUK_SERVER_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			cfg.ServerPort = p
		}
	}

	return cfg, nil
}

// parseConfig parses configuration data from YAML or JSON format.
func parseConfig(data []byte, cfg *Config) error {
	// Default to YAML.
	return yaml.Unmarshal(data, cfg)
}

// VelocityMultiplier returns the multiplier for converting meters per second to the configured velocity unit.
func (c *Config) VelocityMultiplier() float64 {
	switch c.VelocityUnit {
	case "kmh":
		return 3.6
	case "mph":
		return 3.6 * 0.6213712
	case "mps":
		return 1
	default:
		return 3.6
	}
}

// VelocityUnitString returns the unit string for the configured velocity unit.
func (c *Config) VelocityUnitString() string {
	switch c.VelocityUnit {
	case "kmh":
		return "km/h"
	case "mph":
		return "mph"
	case "mps":
		return "m/s"
	default:
		return "km/h"
	}
}

// Duration returns the max duration as a time.Duration.
func (c *Config) Duration() time.Duration {
	return time.Duration(c.MaxDuration) * time.Second
}

// MinIntervalDuration returns the min interval as a time.Duration.
func (c *Config) MinIntervalDuration() time.Duration {
	return time.Duration(c.MinInterval) * time.Second
}

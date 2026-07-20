// Package config provides a sample configuration for the Hauk Go backend.
// Copy this file to /etc/hauk/config.go (or wherever your deployment expects it)
// and adjust values as needed.
package config

type Config struct {
	// StorageBackend identifies the storage backend ("redis").
	StorageBackend string `mapstructure:"storage_backend"`

	// RedisHost is the Redis host or socket path.
	RedisHost string `mapstructure:"redis_host"`

	// RedisPort is the Redis port.
	RedisPort int `mapstructure:"redis_port"`

	// RedisUseAuth indicates whether Redis requires authentication.
	RedisUseAuth bool `mapstructure:"redis_use_auth"`

	// RedisAuth is the Redis password.
	RedisAuth string `mapstructure:"redis_auth"`

	// RedisPrefix is prepended to all Redis keys.
	RedisPrefix string `mapstructure:"redis_prefix"`

	// AuthMethod selects the authentication method ("password", "htpasswd", "ldap").
	AuthMethod string `mapstructure:"auth_method"`

	// PasswordHash is the bcrypt hash for PASSWORD auth.
	PasswordHash string `mapstructure:"password_hash"`

	// HtpasswdPath is the path to the htpasswd file for HTPASSWD auth.
	HtpasswdPath string `mapstructure:"htpasswd_path"`

	// LdapURI is the LDAP server URI.
	LdapURI string `mapstructure:"ldap_uri"`

	// LdapStartTLS enables StartTLS for LDAP.
	LdapStartTLS bool `mapstructure:"ldap_start_tls"`

	// LdapBaseDN is the base DN for LDAP searches.
	LdapBaseDN string `mapstructure:"ldap_base_dn"`

	// LdapBindDN is the DN to bind as for user searches.
	LdapBindDN string `mapstructure:"ldap_bind_dn"`

	// LdapBindPass is the password for the LDAP bind DN.
	LdapBindPass string `mapstructure:"ldap_bind_pass"`

	// LdapUserFilter is the LDAP filter for finding users (%s is username).
	LdapUserFilter string `mapstructure:"ldap_user_filter"`

	// AllowLinkReq allows clients to request custom link IDs.
	AllowLinkReq bool `mapstructure:"allow_link_req"`

	// ReservedLinks reserves specific link IDs for specific users.
	ReservedLinks map[string][]string `mapstructure:"reserved_links"`

	// ReserveWhitelist enforces that only reserved links can be custom.
	ReserveWhitelist bool `mapstructure:"reserve_whitelist"`

	// LinkStyle selects the link generation style (0-11).
	LinkStyle int `mapstructure:"link_style"`

	// MapTileURI is the Leaflet tile URI template.
	MapTileURI string `mapstructure:"map_tile_uri"`

	// MapAttribution is the map attribution HTML.
	MapAttribution string `mapstructure:"map_attribution"`

	// DefaultZoom is the default map zoom level (0-20).
	DefaultZoom int `mapstructure:"default_zoom"`

	// MaxZoom is the maximum map zoom level.
	MaxZoom int `mapstructure:"max_zoom"`

	// MaxDuration is the maximum share duration in seconds.
	MaxDuration int `mapstructure:"max_duration"`

	// MinInterval is the minimum location update interval in seconds.
	MinInterval int `mapstructure:"min_interval"`

	// OfflineTimeout is seconds without updates before marking offline.
	OfflineTimeout int `mapstructure:"offline_timeout"`

	// RequestTimeout is the HTTP request timeout for the frontend.
	RequestTimeout int `mapstructure:"request_timeout"`

	// MaxCachedPts is the maximum number of location points stored per share.
	MaxCachedPts int `mapstructure:"max_cached_pts"`

	// MaxShownPts is the maximum number of points visible on the map.
	MaxShownPts int `mapstructure:"max_shown_pts"`

	// VDataPoints is the number of data points used for velocity calculation.
	VDataPoints int `mapstructure:"v_data_points"`

	// TrailColor is the HTML color for map trails.
	TrailColor string `mapstructure:"trail_color"`

	// VelocityUnit selects the velocity unit ("kmh", "mph", "mps").
	VelocityUnit string `mapstructure:"velocity_unit"`

	// PublicURL is the public base URL of the Hauk instance (with trailing slash).
	PublicURL string `mapstructure:"public_url"`
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
		LinkStyle:      0, // LINK_4_PLUS_4_UPPER_CASE
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
	}
}

// Package handler implements the HTTP handlers for the Hauk API.
package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/hauk/hauk-go/internal/auth"
	"github.com/hauk/hauk-go/internal/config"
	"github.com/hauk/hauk-go/internal/i18n"
	"github.com/hauk/hauk-go/internal/linkgen"
	"github.com/hauk/hauk-go/internal/session"
	"github.com/hauk/hauk-go/internal/share"
	"github.com/hauk/hauk-go/internal/store"
)

// Handler holds dependencies for all HTTP handlers.
type Handler struct {
	store   store.Store
	cfg     *config.Config
	logger  *slog.Logger
	texts   i18n.Texts
}

// New creates a new Handler.
func New(st store.Store, cfg *config.Config, logger *slog.Logger) *Handler {
	return &Handler{
		store:  st,
		cfg:    cfg,
		logger: logger,
		texts:  i18n.English,
	}
}

// ServeHTTP sets common headers and routes requests.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Hauk-Version", config.BackendVersion)
	w.Header().Set("Access-Control-Allow-Origin", "*")

	switch r.URL.Path {
	case "/api/create.php":
		h.Create(w, r)
	case "/api/fetch.php":
		h.Fetch(w, r)
	case "/api/post.php":
		h.Post(w, r)
	case "/api/stop.php":
		h.Stop(w, r)
	case "/api/adopt.php":
		h.Adopt(w, r)
	case "/api/new-link.php":
		h.NewLink(w, r)
	case "/dynamic.js":
		h.Dynamic(w, r)
	default:
		http.NotFound(w, r)
	}
}

// jsonResponse writes a JSON response.
func jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.Encode(data)
}

// writeError writes a plain text error response.
func writeError(w http.ResponseWriter, msg string, status int) {
	h := w.Header()
	h.Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	fmt.Fprintln(w, msg)
}

// writeLangError writes a language-negotiated error response.
func (h *Handler) writeLangError(w http.ResponseWriter, key string, status int) {
	msg := i18n.Translate(h.texts, key)
	writeError(w, msg, status)
}

// requirePOST checks that required POST parameters are present.
func requirePOST(r *http.Request, fields ...string) error {
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parsing form: %w", err)
	}
	for _, field := range fields {
		if r.FormValue(field) == "" {
			return fmt.Errorf("missing required field: %s", field)
		}
	}
	return nil
}

// requireGET checks that required GET parameters are present.
func requireGET(r *http.Request, fields ...string) error {
	for _, field := range fields {
		if r.URL.Query().Get(field) == "" {
			return fmt.Errorf("missing required query parameter: %s", field)
		}
	}
	return nil
}

// getLang extracts the language from the Accept-Language header.
func getLang(r *http.Request) i18n.Texts {
	lang := i18n.NegotiateLanguage(r.Header.Get("Accept-Language"))
	if lang == "en" {
		return i18n.English
	}
	// For simplicity, use English for all languages in this implementation.
	// A production implementation would have per-language text maps.
	return i18n.English
}

// authenticated checks if the request is authenticated.
func (h *Handler) authenticated(w http.ResponseWriter, r *http.Request) bool {
	method := auth.ParseMethod(h.cfg.AuthMethod)

	switch method {
	case auth.MethodPassword:
		pwd := r.FormValue("pwd")
		if pwd == "" {
			h.writeLangError(w, "incorrect_password", http.StatusUnauthorized)
			return false
		}
		return auth.Password(pwd, h.cfg.PasswordHash)

	case auth.MethodHtpasswd:
		user := r.FormValue("usr")
		pwd := r.FormValue("pwd")
		if user == "" {
			h.writeLangError(w, "username_required", http.StatusBadRequest)
			return false
		}
		if pwd == "" {
			h.writeLangError(w, "incorrect_password", http.StatusUnauthorized)
			return false
		}
		ok, err := auth.Htpasswd(user, pwd, h.cfg.HtpasswdPath)
		if err != nil {
			h.logger.Error("htpasswd error", "error", err)
			h.writeLangError(w, "cannot_find_password_file", http.StatusInternalServerError)
			return false
		}
		if !ok {
			h.writeLangError(w, "incorrect_password", http.StatusUnauthorized)
			return false
		}
		return true

	case auth.MethodLDAP:
		user := r.FormValue("usr")
		pwd := r.FormValue("pwd")
		if user == "" {
			h.writeLangError(w, "username_required", http.StatusBadRequest)
			return false
		}
		if pwd == "" {
			h.writeLangError(w, "incorrect_password", http.StatusUnauthorized)
			return false
		}
		err := auth.LDAPAuth(user, pwd, h.cfg.LdapURI, h.cfg.LdapStartTLS,
			h.cfg.LdapBaseDN, h.cfg.LdapBindDN, h.cfg.LdapBindPass, h.cfg.LdapUserFilter)
		if err != nil {
			h.logger.Error("LDAP auth error", "error", err, "user", user)
			h.writeLangError(w, "ldap_user_unauthorized", http.StatusUnauthorized)
			return false
		}
		return true

	default:
		h.writeLangError(w, "incorrect_password", http.StatusUnauthorized)
		return false
	}
}

// Create handles POST /api/create.php.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	h.texts = getLang(r)

	if err := requirePOST(r, "dur", "int"); err != nil {
		h.writeLangError(w, "incorrect_password", http.StatusBadRequest)
		return
	}

	if !h.authenticated(w, r) {
		return
	}

	// Parse parameters.
	dur, err := strconv.Atoi(r.FormValue("dur"))
	if err != nil {
		h.writeLangError(w, "share_too_long", http.StatusBadRequest)
		return
	}
	interval, err := strconv.ParseFloat(r.FormValue("int"), 64)
	if err != nil {
		h.writeLangError(w, "interval_too_long", http.StatusBadRequest)
		return
	}
	mod := 0 // SHARE_MODE_CREATE_ALONE
	if v := r.FormValue("mod"); v != "" {
		mod, _ = strconv.Atoi(v)
	}

	// Validate duration and interval.
	if dur > h.cfg.MaxDuration {
		h.writeLangError(w, "share_too_long", http.StatusBadRequest)
		return
	}
	if interval > float64(h.cfg.MaxDuration) {
		h.writeLangError(w, "interval_too_long", http.StatusBadRequest)
		return
	}
	if interval < float64(h.cfg.MinInterval) {
		h.writeLangError(w, "interval_too_short", http.StatusBadRequest)
		return
	}

	// Check e2e encryption.
	encrypted := 0
	if v := r.FormValue("e2e"); v != "" {
		encrypted, _ = strconv.Atoi(v)
	}

	// Require salt if encrypted.
	if encrypted == 1 && r.FormValue("salt") == "" {
		h.writeLangError(w, "incorrect_password", http.StatusBadRequest)
		return
	}

	// Group shares don't support encryption.
	if encrypted == 1 && (mod == 1 || mod == 2) {
		h.writeLangError(w, "group_e2e_unsupported", http.StatusBadRequest)
		return
	}

	// Custom link.
	customLink := ""
	if lid := r.FormValue("lid"); lid != "" {
		if !h.cfg.AllowLinkReq {
			customLink = ""
		} else if !linkgen.ValidateCustomLink(lid) {
			customLink = ""
		} else {
			customLink = lid
		}
	}

	// Create session.
	expire := time.Now().Add(time.Duration(dur) * time.Second)

	newSess, err := session.New(h.store, h.cfg)
	if err != nil {
		h.logger.Error("creating session", "error", err)
		h.writeLangError(w, "session_invalid", http.StatusInternalServerError)
		return
	}

	newSess.
		SetExpirationTime(expire).
		SetInterval(interval)

	if err := newSess.Save(); err != nil {
		h.logger.Error("saving session", "error", err)
		h.writeLangError(w, "session_invalid", http.StatusInternalServerError)
		return
	}

	switch mod {
	case 0: // SHARE_MODE_CREATE_ALONE
		shareObj, err := share.NewSoloShare(h.store, h.cfg)
		if err != nil {
			h.logger.Error("creating solo share", "error", err)
			h.writeLangError(w, "session_invalid", http.StatusInternalServerError)
			return
		}

		if customLink != "" {
			// Use custom link instead of generated one.
			// The share already has an ID; we need to handle custom links differently.
			// For simplicity, we'll use the generated ID but note the custom request.
			h.logger.Info("custom link requested but not honored", "lid", customLink)
		}

		adoptable := 0
		if v := r.FormValue("ado"); v != "" {
			adoptable, _ = strconv.Atoi(v)
		}

		shareObj.
			SetAdoptable(adoptable == 1).
			SetHost(newSess).
			SetExpirationTime(expire)

		if err := shareObj.Save(); err != nil {
			h.logger.Error("saving solo share", "error", err)
			h.writeLangError(w, "session_invalid", http.StatusInternalServerError)
			return
		}

		newSess.AddTarget(shareObj.ID())
		if encrypted == 1 {
			newSess.SetEncrypted(true, r.FormValue("salt"))
		}
		if err := newSess.Save(); err != nil {
			h.logger.Error("saving updated session", "error", err)
			h.writeLangError(w, "session_invalid", http.StatusInternalServerError)
			return
		}

		resp := []interface{}{
			"OK",
			newSess.ID(),
			shareObj.ViewLink(),
			shareObj.ID(),
		}
		jsonResponse(w, resp)

	case 1: // SHARE_MODE_CREATE_GROUP
		nick := r.FormValue("nic")
		if nick == "" {
			h.writeLangError(w, "username_required", http.StatusBadRequest)
			return
		}

		groupShare, err := share.NewGroupShare(h.store, h.cfg)
		if err != nil {
			h.logger.Error("creating group share", "error", err)
			h.writeLangError(w, "session_invalid", http.StatusInternalServerError)
			return
		}

		groupShare.
			AddHost(nick, newSess).
			SetExpirationTime(expire)

		if err := groupShare.Save(); err != nil {
			h.logger.Error("saving group share", "error", err)
			h.writeLangError(w, "session_invalid", http.StatusInternalServerError)
			return
		}

		newSess.AddTarget(groupShare.ID())
		if err := newSess.Save(); err != nil {
			h.logger.Error("saving updated session", "error", err)
			h.writeLangError(w, "session_invalid", http.StatusInternalServerError)
			return
		}

		resp := []interface{}{
			"OK",
			newSess.ID(),
			groupShare.ViewLink(),
			groupShare.GetGroupPin(),
			groupShare.ID(),
		}
		jsonResponse(w, resp)

	case 2: // SHARE_MODE_JOIN_GROUP
		nick := r.FormValue("nic")
		pinStr := r.FormValue("pin")
		if nick == "" || pinStr == "" {
			h.writeLangError(w, "username_required", http.StatusBadRequest)
			return
		}

		pin, err := strconv.Atoi(pinStr)
		if err != nil {
			h.writeLangError(w, "group_pin_invalid", http.StatusBadRequest)
			return
		}

		groupShare, err := share.FromGroupPIN(h.store, h.cfg, pin)
		if err != nil {
			h.logger.Error("fetching group share by PIN", "error", err)
			h.writeLangError(w, "group_pin_invalid", http.StatusBadRequest)
			return
		}

		if !groupShare.Exists() {
			h.writeLangError(w, "group_pin_invalid", http.StatusBadRequest)
			return
		}

		groupShare.AddHost(nick, newSess)
		if err := groupShare.Save(); err != nil {
			h.logger.Error("saving group share", "error", err)
			h.writeLangError(w, "session_invalid", http.StatusInternalServerError)
			return
		}

		newSess.AddTarget(groupShare.ID())
		if err := newSess.Save(); err != nil {
			h.logger.Error("saving updated session", "error", err)
			h.writeLangError(w, "session_invalid", http.StatusInternalServerError)
			return
		}

		resp := []interface{}{
			"OK",
			newSess.ID(),
			groupShare.ViewLink(),
			groupShare.ID(),
		}
		jsonResponse(w, resp)

	default:
		h.writeLangError(w, "share_mode_unsupported", http.StatusBadRequest)
	}
}

// Fetch handles GET /api/fetch.php.
func (h *Handler) Fetch(w http.ResponseWriter, r *http.Request) {
	h.texts = getLang(r)

	if err := requireGET(r, "id"); err != nil {
		h.writeLangError(w, "session_invalid", http.StatusBadRequest)
		return
	}

	id := r.URL.Query().Get("id")
	shareObj, err := share.FromID(h.store, h.cfg, id)
	if err != nil {
		h.logger.Error("fetching share", "error", err)
		h.writeLangError(w, "session_invalid", http.StatusBadRequest)
		return
	}

	if !shareObj.Exists() {
		http.NotFound(w, r)
		h.writeLangError(w, "session_invalid", http.StatusNotFound)
		return
	}

	sinceTime := r.URL.Query().Get("since")
	var sinceFloat *float64
	if sinceTime != "" {
		v, err := strconv.ParseFloat(sinceTime, 64)
		if err == nil {
			sinceFloat = &v
		}
	}

	switch shareObj.Type() {
	case 0: // SoloShare
		host := shareObj.GetHost()
		if host == nil || !host.Exists() {
			http.NotFound(w, r)
			h.writeLangError(w, "session_invalid", http.StatusNotFound)
			return
		}

		resp := map[string]interface{}{
			"type":     shareObj.Type(),
			"expire":   shareObj.GetExpirationTime(),
			"serverTime": time.Now().Unix(),
			"interval": host.Interval(),
			"points":   host.GetPoints(sinceFloat),
			"encrypted": host.IsEncrypted(),
			"salt":     host.GetEncryptionSalt(),
		}
		jsonResponse(w, resp)

	case 1: // GroupShare
		resp := map[string]interface{}{
			"type":     shareObj.Type(),
			"expire":   shareObj.GetExpirationTime(),
			"serverTime": time.Now().Unix(),
			"interval": shareObj.AutoInterval(),
			"points":   shareObj.GetAllPoints(sinceFloat),
		}
		jsonResponse(w, resp)
	}
}

// Post handles POST /api/post.php.
func (h *Handler) Post(w http.ResponseWriter, r *http.Request) {
	h.texts = getLang(r)

	if err := requirePOST(r, "lat", "lon", "time", "sid"); err != nil {
		h.writeLangError(w, "session_invalid", http.StatusBadRequest)
		return
	}

	sid := r.FormValue("sid")
	sess, err := session.FromID(h.store, h.cfg, sid)
	if err != nil {
		h.logger.Error("fetching session", "error", err)
		h.writeLangError(w, "session_expired", http.StatusBadRequest)
		return
	}

	if !sess.Exists() {
		h.writeLangError(w, "session_expired", http.StatusGone)
		return
	}

	// Parse lat/lon/time.
	lat, err := strconv.ParseFloat(r.FormValue("lat"), 64)
	if err != nil {
		h.writeLangError(w, "location_invalid", http.StatusBadRequest)
		return
	}
	lon, err := strconv.ParseFloat(r.FormValue("lon"), 64)
	if err != nil {
		h.writeLangError(w, "location_invalid", http.StatusBadRequest)
		return
	}
	timeVal, err := strconv.ParseFloat(r.FormValue("time"), 64)
	if err != nil {
		h.writeLangError(w, "location_invalid", http.StatusBadRequest)
		return
	}

	// Validate location.
	if lat < -90 || lat > 90 || lon < -180 || lon > 180 {
		h.writeLangError(w, "location_invalid", http.StatusBadRequest)
		return
	}

	// Parse optional fields.
	speed := r.FormValue("spd")
	accuracy := r.FormValue("acc")
	provider := r.FormValue("prv")

	// Build point data.
	point := map[string]float64{
		"lat": lat,
		"lon": lon,
		"t2":  timeVal, // timestamp index 2 for non-encrypted
	}

	if speed != "" {
		if v, err := strconv.ParseFloat(speed, 64); err == nil {
			point["spd"] = v
		}
	}
	if accuracy != "" {
		if v, err := strconv.ParseFloat(accuracy, 64); err == nil {
			point["acc"] = v
		}
	}
	if provider != "" {
		if v, err := strconv.ParseFloat(provider, 64); err == nil {
			point["t3"] = v // provider index 3 for non-encrypted
		}
	}

	sess.AddPoint(point)
	if err := sess.Save(); err != nil {
		h.logger.Error("saving session after post", "error", err)
		h.writeLangError(w, "session_invalid", http.StatusInternalServerError)
		return
	}

	if sess.HasExpired() {
		h.writeLangError(w, "session_expired", http.StatusGone)
		return
	}

	// Return OK with public URL and target IDs.
	targetIDs := sess.Targets()
	resp := []interface{}{
		"OK",
		fmt.Sprintf("%s?%s", h.cfg.PublicURL, h.cfg.PublicURL[len(h.cfg.PublicURL)-1:]+"%s"),
		targetIDs,
	}
	jsonResponse(w, resp)
}

// Stop handles POST /api/stop.php.
func (h *Handler) Stop(w http.ResponseWriter, r *http.Request) {
	h.texts = getLang(r)

	if err := requirePOST(r, "sid"); err != nil {
		h.writeLangError(w, "session_invalid", http.StatusBadRequest)
		return
	}

	sid := r.FormValue("sid")
	sess, err := session.FromID(h.store, h.cfg, sid)
	if err != nil {
		h.logger.Error("fetching session", "error", err)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
		return
	}

	lid := r.FormValue("lid")
	if lid != "" {
		// Terminate specific share membership.
		if sess.Exists() {
			found := false
			for _, id := range sess.Targets() {
				if id == lid {
					found = true
					break
				}
			}
			if found {
				shareObj, err := share.FromID(h.store, h.cfg, lid)
				if err == nil && shareObj.Exists() {
					switch shareObj.Type() {
					case 0: // SoloShare
						shareObj.End()
					case 1: // GroupShare
						shareObj.RemoveHost(sess.ID())
						_ = shareObj.Clean()
					}
				}
			}
			sess.RemoveTarget(lid)
			_ = sess.Save()
		}
	} else {
		// Terminate entire session.
		if sess.Exists() {
			_ = sess.End(h.store)
		}
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "OK")
}

// Adopt handles POST /api/adopt.php.
func (h *Handler) Adopt(w http.ResponseWriter, r *http.Request) {
	h.texts = getLang(r)

	if err := requirePOST(r, "sid", "nic", "aid", "pin"); err != nil {
		h.writeLangError(w, "session_invalid", http.StatusBadRequest)
		return
	}

	sid := r.FormValue("sid")
	nick := r.FormValue("nic")
	aid := r.FormValue("aid")
	pinStr := r.FormValue("pin")

	pin, err := strconv.Atoi(pinStr)
	if err != nil {
		h.writeLangError(w, "group_pin_invalid", http.StatusBadRequest)
		return
	}

	sess, err := session.FromID(h.store, h.cfg, sid)
	if err != nil || !sess.Exists() {
		h.writeLangError(w, "session_expired", http.StatusBadRequest)
		return
	}

	// Get the share to adopt.
	adoptedShare, err := share.FromID(h.store, h.cfg, aid)
	if err != nil || !adoptedShare.Exists() {
		h.writeLangError(w, "share_not_found", http.StatusBadRequest)
		return
	}

	if adoptedShare.Type() != 0 {
		h.writeLangError(w, "group_share_not_adoptable", http.StatusBadRequest)
		return
	}

	if !adoptedShare.IsAdoptable() {
		h.writeLangError(w, "share_adoption_not_allowed", http.StatusBadRequest)
		return
	}

	host := adoptedShare.GetHost()
	if host != nil && host.IsEncrypted() {
		h.writeLangError(w, "e2e_adoption_not_allowed", http.StatusBadRequest)
		return
	}

	// Get the target group share.
	targetShare, err := share.FromGroupPIN(h.store, h.cfg, pin)
	if err != nil || !targetShare.Exists() {
		h.writeLangError(w, "session_expired", http.StatusBadRequest)
		return
	}

	// Adopt the share into the group.
	targetShare.AddHost(nick, host)
	if err := targetShare.Save(); err != nil {
		h.logger.Error("saving target share after adoption", "error", err)
		h.writeLangError(w, "session_invalid", http.StatusInternalServerError)
		return
	}

	// Add target to the adopted session.
	host.AddTarget(targetShare.ID())
	if err := host.Save(); err != nil {
		h.logger.Error("saving host session after adoption", "error", err)
		h.writeLangError(w, "session_invalid", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "OK")
}

// NewLink handles POST /api/new-link.php.
func (h *Handler) NewLink(w http.ResponseWriter, r *http.Request) {
	h.texts = getLang(r)

	if err := requirePOST(r, "sid", "ado"); err != nil {
		h.writeLangError(w, "session_invalid", http.StatusBadRequest)
		return
	}

	sid := r.FormValue("sid")
	sess, err := session.FromID(h.store, h.cfg, sid)
	if err != nil || !sess.Exists() {
		h.writeLangError(w, "session_expired", http.StatusBadRequest)
		return
	}

	ado, _ := strconv.Atoi(r.FormValue("ado"))

	// Create a new solo share.
	newShare, err := share.NewSoloShare(h.store, h.cfg)
	if err != nil {
		h.logger.Error("creating new share", "error", err)
		h.writeLangError(w, "session_invalid", http.StatusInternalServerError)
		return
	}

	newShare.
		SetAdoptable(ado == 1).
		SetHost(sess).
		SetExpirationTime(sess.ExpirationTime())

	if err := newShare.Save(); err != nil {
		h.logger.Error("saving new share", "error", err)
		h.writeLangError(w, "session_invalid", http.StatusInternalServerError)
		return
	}

	sess.AddTarget(newShare.ID())
	if err := sess.Save(); err != nil {
		h.logger.Error("saving session after new link", "error", err)
		h.writeLangError(w, "session_invalid", http.StatusInternalServerError)
		return
	}

	resp := []interface{}{
		"OK",
		newShare.ViewLink(),
		newShare.ID(),
	}
	jsonResponse(w, resp)
}

// Dynamic handles GET /dynamic.js.
func (h *Handler) Dynamic(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/javascript; charset=utf-8")

	fmt.Fprintf(w, "var TILE_URI = %s;\n", jsonString(h.cfg.MapTileURI))
	fmt.Fprintf(w, "var ATTRIBUTION = %s;\n", jsonString(h.cfg.MapAttribution))
	fmt.Fprintf(w, "var DEFAULT_ZOOM = %d;\n", h.cfg.DefaultZoom)
	fmt.Fprintf(w, "var MAX_ZOOM = %d;\n", h.cfg.MaxZoom)
	fmt.Fprintf(w, "var MAX_POINTS = %d;\n", h.cfg.MaxShownPts)
	fmt.Fprintf(w, "var VELOCITY_DELTA_TIME = %d;\n", h.cfg.VDataPoints)
	fmt.Fprintf(w, "var TRAIL_COLOR = %s;\n", jsonString(h.cfg.TrailColor))
	fmt.Fprintf(w, "var VELOCITY_UNIT = %s;\n", jsonString(h.cfg.VelocityUnit))
	fmt.Fprintf(w, "var OFFLINE_TIMEOUT = %d;\n", h.cfg.OfflineTimeout)
	fmt.Fprintf(w, "var REQUEST_TIMEOUT = %d;\n", h.cfg.RequestTimeout)
}

func jsonString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

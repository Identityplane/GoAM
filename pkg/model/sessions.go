package model

import "time"

// DeviceAttributeValue is the attribute value for devices
// @description Device information
type DeviceAttributeValue struct {
	DeviceID         string `json:"device_id" example:"1234567890"`
	DeviceSecretHash string `json:"device_secret_hash" example:"1234567890"`
	DeviceName       string `json:"device_name" example:"John Doe's iPhone"`
	DeviceType       string `json:"device_type" example:"mobile"`
	DeviceOS         string `json:"device_os" example:"iOS"`
	DeviceOSVersion  string `json:"device_os_version" example:"15.0"`
	DeviceModel      string `json:"device_model" example:"iPhone 12"`
	DeviceIP         string `json:"device_ip" example:"192.168.1.100"`
	DeviceUserAgent  string `json:"device_user_agent" example:"Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.0 Mobile/15E148 Safari/604.1"`

	CookieName     string    `json:"cookie_name" example:"session"`
	CookieExpires  time.Time `json:"cookie_expires" example:"2024-01-01T00:00:00Z"`
	CookieSameSite string    `json:"cookie_same_site" example:"Lax"`
	CookieHttpOnly bool      `json:"cookie_http_only" example:"true"`
	CookieSecure   bool      `json:"cookie_secure" example:"true"`

	SessionLoa0 Session  `json:"session_loa0" example:"{}"` // Session for LOA0 (typically no active authentication e.g. device only)
	SessionLoa1 *Session `json:"session_loa1" example:"{}"` // Session for LOA1 (typically normal authentication)
	SessionLoa2 *Session `json:"session_loa2" example:"{}"` // Session for LOA2 (typically strong authentication)
}

// Session is a session for a device
type Session struct {
	FirstLoginTime      time.Time `json:"first_login_time"`      // The first time the user was authenticated using this session
	LastLoginTime       time.Time `json:"last_login_time"`       // The last time the user was authenticated using this session
	SessionLastActivity time.Time `json:"session_last_activity"` // The last this session was active (a API call was made or a page was loaded)

	SessionDuration     int       `json:"session_duration" example:"3600"`               // The duration of the session in seconds
	SessionExpiry       time.Time `json:"session_expiry" example:"2024-01-01T00:00:00Z"` // The expiry time of the session
	SessionRefreshAfter int       `json:"session_refresh_after" example:"1800"`          // The time after which the session will be refreshed (if the session has an activity after expiry-refresh but before expiry, the session will be refreshed)

	LevelOfAssurance int `json:"level_of_assurance"` // The level of assurance for the session
}

// LoaToExpiryMapping is a mapping of level of assurance to duration and refresh after
type LoaToExpiryMapping struct {
	Loa          int `json:"loa"`
	Duration     int `json:"duration"`
	RefreshAfter int `json:"refresh_after"`
}

// IsActive returns true if the session is active
func (s *Session) IsActive(now time.Time) bool {
	return s.SessionExpiry.After(now)
}

// ShouldRefresh returns true if the session should be refreshed
func (s *Session) ShouldRefresh(now time.Time) bool {

	if s == nil {
		return false
	}

	if s.SessionRefreshAfter < 0 {
		return false
	}
	return s.IsActive(now) && s.SessionExpiry.Add(-time.Duration(s.SessionDuration-s.SessionRefreshAfter)*time.Second).Before(now)
}

// Init initializes the session
func InitSession(now time.Time, mapping LoaToExpiryMapping) *Session {
	return &Session{
		FirstLoginTime:      now,
		LastLoginTime:       now,
		SessionLastActivity: now,
		SessionDuration:     mapping.Duration,
		SessionExpiry:       now.Add(time.Duration(mapping.Duration) * time.Second),
		SessionRefreshAfter: mapping.RefreshAfter,
		LevelOfAssurance:    mapping.Loa,
	}
}

// Refresh refreshes the session if it should be refreshed
func (s *Session) Refresh(now time.Time) {

	if !s.ShouldRefresh(now) {
		return
	}

	s.SessionExpiry = now.Add(time.Duration(s.SessionDuration) * time.Second)
	s.SessionLastActivity = now
}

// Refresh refreshes all active sessions of the device
func (d *DeviceAttributeValue) Refresh(now time.Time) {
	d.SessionLoa0.Refresh(now)
	d.SessionLoa1.Refresh(now)
	d.SessionLoa2.Refresh(now)
}

// LatestExpiry returns the latest expiry time of all active sessions
func (d *DeviceAttributeValue) LatestExpiry(now time.Time) time.Time {

	expiry := d.SessionLoa0.SessionExpiry
	if d.SessionLoa1 != nil && d.SessionLoa1.IsActive(now) && d.SessionLoa1.SessionExpiry.After(expiry) {
		expiry = d.SessionLoa1.SessionExpiry
	}
	if d.SessionLoa2 != nil && d.SessionLoa2.IsActive(now) && d.SessionLoa2.SessionExpiry.After(expiry) {
		expiry = d.SessionLoa2.SessionExpiry
	}
	return expiry
}

// GetLatestAuthTime returns the latest authentication time of all active sessions
func (d *DeviceAttributeValue) GetLatestAuthTime(now time.Time) time.Time {
	authTime := d.SessionLoa0.FirstLoginTime
	if d.SessionLoa1 != nil && d.SessionLoa1.IsActive(now) && d.SessionLoa1.FirstLoginTime.After(authTime) {
		authTime = d.SessionLoa1.FirstLoginTime
	}
	if d.SessionLoa2 != nil && d.SessionLoa2.IsActive(now) && d.SessionLoa2.FirstLoginTime.After(authTime) {
		authTime = d.SessionLoa2.FirstLoginTime
	}
	return authTime
}

// CurrentLoa returns the highest level of assurance of all active sessions
func (d *DeviceAttributeValue) CurrentLoa(now time.Time) int {
	loa := d.SessionLoa0.LevelOfAssurance
	if d.SessionLoa1 != nil && d.SessionLoa1.IsActive(now) && d.SessionLoa1.LevelOfAssurance > loa {
		loa = d.SessionLoa1.LevelOfAssurance
	}
	if d.SessionLoa2 != nil && d.SessionLoa2.IsActive(now) && d.SessionLoa2.LevelOfAssurance > loa {
		loa = d.SessionLoa2.LevelOfAssurance
	}
	return loa
}

// DEFAULT_LOA_TO_EXPIRY_MAPPINGS is the default mapping of level of assurance to duration and refresh after
var DEFAULT_LOA_TO_EXPIRY_MAPPINGS = []LoaToExpiryMapping{
	{
		Loa:          0,               // LOA0 typically no active authentication e.g. device only
		Duration:     3600 * 24 * 365, // 1 year
		RefreshAfter: 3600 * 24 * 30,  // 30 days
	},
	{
		Loa:          1,        // LOA1 typically normal authentication
		Duration:     3600 * 2, // 2 hours
		RefreshAfter: 3600,     // 1 hour
	},
	{
		Loa:          2,   // LOA2 typically strong authentication
		Duration:     300, // 5 minutes
		RefreshAfter: -1,  // No refresh
	},
}

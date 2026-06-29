package auth

// Credentials holds the LeetCode session credentials needed for some LeetCode operations.
// The 'toml' tags specify how the fields are serialized/deserialized in TOML format, can be ignored
// in most cases.
type Credentials struct {
	SessionToken string `toml:"session_token"`
	CSRFToken    string `toml:"csrf_token"`
}

// IsSet returns true if both the SessionToken and CSRFToken are set (non-empty), since having only
// one without the other is not enough.
func (c Credentials) IsSet() bool {
	return c.SessionToken != "" && c.CSRFToken != ""
}

// IsEqual checks if two Credentials instances are equal.
func (c Credentials) IsEqual(other Credentials) bool {
	return c.SessionToken == other.SessionToken && c.CSRFToken == other.CSRFToken
}

package aggregate

// JwtToken represents a JWT token pair (access + refresh)
type JwtToken struct {
	accessToken  string
	refreshToken string
	roles        []string
}

// NewJwtToken creates a new JwtToken
func NewJwtToken() *JwtToken {
	return &JwtToken{}
}

// SetAccessToken sets the access token
func (j *JwtToken) SetAccessToken(accessToken string) {
	j.accessToken = accessToken
}

// SetRefreshToken sets the refresh token
func (j *JwtToken) SetRefreshToken(refreshToken string) {
	j.refreshToken = refreshToken
}

// SetRoles sets the roles
func (j *JwtToken) SetRoles(roles []string) {
	j.roles = roles
}

// GetAccessToken returns the access token
func (j *JwtToken) GetAccessToken() string {
	return j.accessToken
}

// GetRefreshToken returns the refresh token
func (j *JwtToken) GetRefreshToken() string {
	return j.refreshToken
}

// GetRoles returns the roles
func (j *JwtToken) GetRoles() []string {
	return j.roles
}

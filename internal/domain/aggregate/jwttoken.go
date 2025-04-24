package aggregate

type JwtToken struct {
	accessToken  string `json:"access_token"`
	refreshToken string `json:"refresh_token"`
}

func (j *JwtToken) SetAccessToken(accessToken string) {
	j.accessToken = accessToken
}

func (j *JwtToken) SetRefreshToken(refreshToken string) {
	j.refreshToken = refreshToken
}

func (j *JwtToken) GetAccessToken() string {
	return j.accessToken
}

func (j *JwtToken) GetRefreshToken() string {
	return j.refreshToken
}

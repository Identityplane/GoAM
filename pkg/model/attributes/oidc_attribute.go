package attributes

type OidcAttributeValue struct {

	// @description The issuer of the oidc provider
	Issuer string `json:"issuer" example:"https://accounts.google.com"`

	// @description The client id of the oidc provider that was used to login
	ClientId string `json:"client_id" example:"1234567890"`

	// @description The scope of the oidc login
	Scope string `json:"scope" example:"openid email profile"`

	// @description The access token from the oidc login
	AccessToken string `json:"access_token" example:"1234567890"`

	// @description The id token from the oidc login
	IDToken string `json:"id_token" example:"1234567890"`

	// @description The refresh token from the oidc login
	RefreshToken string `json:"refresh_token" example:"1234567890"`

	// @description The expires in of the access token
	ExpiresIn int64 `json:"expires_in" example:"3600"`

	// @description The subject of the oidc login
	Sub string `json:"sub" example:"1234567890"`

	// @description The claims from the oidc login
	Claims map[string]interface{} `json:"claims" example:"map[string]interface{}"`
}

// GetIndex returns the index of the oidc attribute value
func (o *OidcAttributeValue) GetIndex() string {
	return o.Issuer + "/" + o.Sub
}

package attributes

// GitHubAttributeValue is the attribute value for GitHub
// @description GitHub information
type GitHubAttributeValue struct {
	GitHubUserID       string `json:"github_user_id" example:"1234567890"`
	GitHubRefreshToken string `json:"github_refresh_token" example:"1234567890"`
	GitHubEmail        string `json:"github_email" example:"john.doe@example.com"`
	GitHubAvatarURL    string `json:"github_avatar_url" example:"https://example.com/avatar.jpg"`
	GitHubUsername     string `json:"github_username" example:"john.doe"`
	GitHubAccessToken  string `json:"github_access_token" example:"1234567890"`
	GitHubTokenType    string `json:"github_token_type" example:"bearer"`
	GitHubScope        string `json:"github_scope" example:"user:email"`
}

// GetIndex returns the index of the GitHub attribute value
func (g *GitHubAttributeValue) GetIndex() string {
	return g.GitHubUserID
}

// IndexIsSensitive returns whether the index should be omitted from JSON API responses
func (g *GitHubAttributeValue) IndexIsSensitive() bool {
	return false // GitHub user ID is not sensitive
}

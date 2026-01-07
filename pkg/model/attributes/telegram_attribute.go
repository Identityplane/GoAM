package attributes

// TelegramAttributeValue is the attribute value for Telegram accounts
// @description Telegram information
type TelegramAttributeValue struct {
	TelegramUserID    string `json:"telegram_user_id" example:"1234567890"`
	TelegramUsername  string `json:"telegram_username" example:"johndoe"`
	TelegramFirstName string `json:"telegram_first_name" example:"John"`
	TelegramPhotoURL  string `json:"telegram_photo_url" example:"https://t.me/i/userpic/123/photo.jpg"`
	TelegramAuthDate  int64  `json:"telegram_auth_date" example:"1753278987"`
}

// GetIndex returns the index of the Telegram attribute value
func (t *TelegramAttributeValue) GetIndex() string {
	return t.TelegramUserID
}

// IndexIsSensitive returns whether the index should be omitted from JSON API responses
func (t *TelegramAttributeValue) IndexIsSensitive() bool {
	return false // Telegram user ID is not sensitive
}

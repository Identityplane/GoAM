package service

import (
	"context"
	"encoding/json"

	"github.com/Identityplane/GoAM/pkg/db"
	"github.com/Identityplane/GoAM/pkg/model"
	"github.com/google/uuid"

	services_interface "github.com/Identityplane/GoAM/pkg/services"
)

// userAttributeServiceImpl implements UserAttributeService
type userAttributeServiceImpl struct {
	userAttributeDB db.UserAttributeDB
	userDB          db.UserDB
}

// NewUserAttributeService creates a new UserAttributeService instance
func NewUserAttributeService(userAttributeDB db.UserAttributeDB, userDB db.UserDB) services_interface.UserAttributeService {
	return &userAttributeServiceImpl{
		userAttributeDB: userAttributeDB,
		userDB:          userDB,
	}
}

//

func (s *userAttributeServiceImpl) ListUserAttributes(ctx context.Context, tenant, realm, userID string) ([]*model.UserAttribute, error) {
	// Verify user exists
	user, err := s.userDB.GetUserByID(ctx, tenant, realm, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil // User not found
	}

	return s.userAttributeDB.ListUserAttributes(ctx, tenant, realm, userID)
}

func (s *userAttributeServiceImpl) GetUserAttributeByID(ctx context.Context, tenant, realm, attributeID string) (*model.UserAttribute, error) {
	return s.userAttributeDB.GetUserAttributeByID(ctx, tenant, realm, attributeID)
}

func (s *userAttributeServiceImpl) CreateUserAttribute(ctx context.Context, attribute model.UserAttribute) (*model.UserAttribute, error) {
	// Verify user exists
	user, err := s.userDB.GetUserByID(ctx, attribute.Tenant, attribute.Realm, attribute.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil // User not found
	}

	// Generate UUID for the new attribute
	if attribute.ID == "" {
		attribute.ID = uuid.NewString()
	}

	// Automatically set index from attribute value if it implements AttributeValue interface
	setIndexFromValue(&attribute)

	// Create the attribute
	err = s.userAttributeDB.CreateUserAttribute(ctx, attribute)
	if err != nil {
		return nil, err
	}

	// Return the created attribute (with ID populated)
	return s.userAttributeDB.GetUserAttributeByID(ctx, attribute.Tenant, attribute.Realm, attribute.ID)
}

func (s *userAttributeServiceImpl) UpdateUserAttribute(ctx context.Context, attribute *model.UserAttribute) error {
	// Verify attribute exists
	existing, err := s.userAttributeDB.GetUserAttributeByID(ctx, attribute.Tenant, attribute.Realm, attribute.ID)
	if err != nil {
		return err
	}
	if existing == nil {
		return nil // Attribute not found
	}

	// Automatically set index from attribute value if it implements AttributeValue interface
	// This ensures the index is updated if the value changed
	setIndexFromValue(attribute)

	return s.userAttributeDB.UpdateUserAttribute(ctx, attribute)
}

func (s *userAttributeServiceImpl) DeleteUserAttribute(ctx context.Context, tenant, realm, attributeID string) error {
	return s.userAttributeDB.DeleteUserAttribute(ctx, tenant, realm, attributeID)
}

// setIndexFromValue extracts the index from the attribute value using GetIndex() method
// if the value implements the AttributeValue interface
func setIndexFromValue(attribute *model.UserAttribute) {
	if attribute.Value == nil {
		return
	}

	// Try to convert the value to AttributeValue interface
	if attrValue, ok := attribute.Value.(model.AttributeValue); ok {
		index := attrValue.GetIndex()
		if index != "" {
			attribute.Index = &index
		} else {
			attribute.Index = nil
		}
		return
	}

	// If direct conversion fails, try to convert from map[string]interface{} (for database stored values)
	if mapValue, ok := attribute.Value.(map[string]interface{}); ok {
		// Convert map to JSON and then try to unmarshal to known attribute types
		jsonData, err := json.Marshal(mapValue)
		if err != nil {
			attribute.Index = nil
			return
		}

		// Try each known attribute type based on the Type field
		switch attribute.Type {
		case model.AttributeTypeEmail:
			var val model.EmailAttributeValue
			if err := json.Unmarshal(jsonData, &val); err == nil {
				index := val.GetIndex()
				if index != "" {
					attribute.Index = &index
				} else {
					attribute.Index = nil
				}
			}
		case model.AttributeTypePhone:
			var val model.PhoneAttributeValue
			if err := json.Unmarshal(jsonData, &val); err == nil {
				index := val.GetIndex()
				if index != "" {
					attribute.Index = &index
				} else {
					attribute.Index = nil
				}
			}
		case model.AttributeTypeUsername:
			var val model.UsernameAttributeValue
			if err := json.Unmarshal(jsonData, &val); err == nil {
				index := val.GetIndex()
				if index != "" {
					attribute.Index = &index
				} else {
					attribute.Index = nil
				}
			}
		case model.AttributeTypePassword:
			var val model.PasswordAttributeValue
			if err := json.Unmarshal(jsonData, &val); err == nil {
				index := val.GetIndex()
				if index != "" {
					attribute.Index = &index
				} else {
					attribute.Index = nil
				}
			}
		case model.AttributeTypeTOTP:
			var val model.TOTPAttributeValue
			if err := json.Unmarshal(jsonData, &val); err == nil {
				index := val.GetIndex()
				if index != "" {
					attribute.Index = &index
				} else {
					attribute.Index = nil
				}
			}
		case model.AttributeTypeGitHub:
			var val model.GitHubAttributeValue
			if err := json.Unmarshal(jsonData, &val); err == nil {
				index := val.GetIndex()
				if index != "" {
					attribute.Index = &index
				} else {
					attribute.Index = nil
				}
			}
		case model.AttributeTypeTelegram:
			var val model.TelegramAttributeValue
			if err := json.Unmarshal(jsonData, &val); err == nil {
				index := val.GetIndex()
				if index != "" {
					attribute.Index = &index
				} else {
					attribute.Index = nil
				}
			}
		case model.AttributeTypePasskey:
			var val model.PasskeyAttributeValue
			if err := json.Unmarshal(jsonData, &val); err == nil {
				index := val.GetIndex()
				if index != "" {
					attribute.Index = &index
				} else {
					attribute.Index = nil
				}
			}
		case model.AttributeTypeYubico:
			var val model.YubicoAttributeValue
			if err := json.Unmarshal(jsonData, &val); err == nil {
				index := val.GetIndex()
				if index != "" {
					attribute.Index = &index
				} else {
					attribute.Index = nil
				}
			}
		case model.AttributeTypeEntitlements:
			var val model.EntitlementSetAttributeValue
			if err := json.Unmarshal(jsonData, &val); err == nil {
				index := val.GetIndex()
				if index != "" {
					attribute.Index = &index
				} else {
					attribute.Index = nil
				}
			}
		case model.AttributeTypeOidc:
			var val model.OidcAttributeValue
			if err := json.Unmarshal(jsonData, &val); err == nil {
				index := val.GetIndex()
				if index != "" {
					attribute.Index = &index
				} else {
					attribute.Index = nil
				}
			}
		case model.AttributeTypeDevice:
			var val model.DeviceAttributeValue
			if err := json.Unmarshal(jsonData, &val); err == nil {
				index := val.GetIndex()
				if index != "" {
					attribute.Index = &index
				} else {
					attribute.Index = nil
				}
			}
		}
	}
}

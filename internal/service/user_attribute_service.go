package service

import (
	"context"

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
// if the value implements the AttributeValue interface. Handles both concrete types and
// map[string]interface{} from JSON unmarshaling.
func setIndexFromValue(attribute *model.UserAttribute) {
	if attribute.Value == nil {
		attribute.Index = nil
		return
	}

	// Try to convert the value to AttributeValue interface directly
	if attrValue, ok := attribute.Value.(model.AttributeValue); ok {
		index := attrValue.GetIndex()
		if index != "" {
			attribute.Index = &index
		} else {
			attribute.Index = nil
		}
		return
	}

	// If direct conversion fails, try to convert from map[string]interface{} (for JSON unmarshaled values)
	if mapValue, ok := attribute.Value.(map[string]interface{}); ok {
		// Use the converter map to convert to AttributeValue
		attrValue := model.ConvertMapToAttributeValue(attribute.Type, mapValue)
		if attrValue != nil {
			index := attrValue.GetIndex()
			if index != "" {
				attribute.Index = &index
			} else {
				attribute.Index = nil
			}
			return
		}
	}

	// If we can't convert to AttributeValue, set index to nil
	attribute.Index = nil
}

package service

import (
	"goiam/internal/db"
)

// UserDB defines the interface for user database operations
type UserDB interface {
	db.UserDB
}

// Services holds all service instances
type Services struct {
	UserService  UserAdminService
	RealmService RealmService
	FlowService  FlowService
}

var (
	// Global service registry
	services *Services
)

// InitServices initializes all services with their dependencies
func InitServices(userDB UserDB) *Services {
	services = &Services{
		UserService:  NewUserService(userDB),
		RealmService: NewRealmService(),
		FlowService:  NewFlowService(),
	}
	return services
}

// GetServices returns the global service registry
func GetServices() *Services {
	if services == nil {
		panic("services not initialized - call InitServices first")
	}
	return services
}

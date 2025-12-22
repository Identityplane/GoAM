package db

// DatabaseConnections holds all database connections
type DatabaseConnections struct {
	UserDB          UserDB
	UserAttributeDB UserAttributeDB
	RealmDB         RealmDB
	FlowDB          FlowDB
	ApplicationsDB  ApplicationDB
	ClientSessionDB ClientSessionDB
	SigningKeyDB    SigningKeyDB
	AuthSessionDB   AuthSessionDB
}

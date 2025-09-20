package model

type RealmObject struct {
	Realm  string `json:"realm" yaml:"realm" db:"realm"`
	Tenant string `json:"tenant" yaml:"tenant" db:"tenant"`
}

package models

type GQLClient interface {
	Query(string, interface{}, map[string]interface{}) error
}
package models

type GQLClient interface {
	Query(string, any, map[string]any) error
}
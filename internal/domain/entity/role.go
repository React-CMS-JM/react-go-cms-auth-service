package entity

// Role is a named set of permission names.
type Role struct {
	ID          int
	Name        string
	Description *string
	Permissions []string
}

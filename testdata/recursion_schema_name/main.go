package main

// User represents a user with self-references
type User struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Children []*User `json:"children,omitempty"`
} // @name User

// @title Test API
// @version 1.0
// @description Test API for recursion with schema name
// @BasePath /
func main() {}

// GetUser returns a user
// @Summary Get user
// @Description Get user by ID
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {object} User
// @Router /user [get]
func GetUser() {}

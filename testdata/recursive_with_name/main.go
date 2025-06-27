package main

import "net/http"

// @title Recursive Type with @name Test API
// @version 1.0
// @description This is a test for recursive types with @name annotations
// @BasePath /api/v1

// EntityHierarchyNode represents a node in the entity hierarchy tree.
// Each node contains an entity and its child nodes.
type EntityHierarchyNode struct {
	ID       string                `json:"id"`
	Name     string                `json:"name"`
	Children []*EntityHierarchyNode `json:"children"`
} // @name EntityHierarchyNode

// TreeData represents a complete tree structure
type TreeData struct {
	Root EntityHierarchyNode `json:"root"`
} // @name TreeData

// GetEntityHierarchy returns the entity hierarchy
// @Summary Get entity hierarchy
// @Description Get the complete entity hierarchy tree
// @Tags hierarchy
// @Accept json
// @Produce json
// @Success 200 {object} TreeData
// @Router /hierarchy [get]
func GetEntityHierarchy(w http.ResponseWriter, r *http.Request) {
	// Implementation
}

func main() {
	// Server setup
}
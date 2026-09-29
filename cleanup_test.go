package swag

import (
	"testing"

	"github.com/go-openapi/spec"
	"github.com/stretchr/testify/assert"
)

func TestRemoveUnusedDefinitions(t *testing.T) {
	t.Run("removes unused definitions", func(t *testing.T) {
		swagger := &spec.Swagger{
			SwaggerProps: spec.SwaggerProps{
				Definitions: spec.Definitions{
					"UsedModel": spec.Schema{
						SchemaProps: spec.SchemaProps{
							Type: []string{"object"},
							Properties: map[string]spec.Schema{
								"name": {
									SchemaProps: spec.SchemaProps{
										Type: []string{"string"},
									},
								},
							},
						},
					},
					"UnusedModel": spec.Schema{
						SchemaProps: spec.SchemaProps{
							Type: []string{"object"},
							Properties: map[string]spec.Schema{
								"value": {
									SchemaProps: spec.SchemaProps{
										Type: []string{"string"},
									},
								},
							},
						},
					},
				},
				Paths: &spec.Paths{
					Paths: map[string]spec.PathItem{
						"/test": {
							PathItemProps: spec.PathItemProps{
								Get: &spec.Operation{
									OperationProps: spec.OperationProps{
										Responses: &spec.Responses{
											ResponsesProps: spec.ResponsesProps{
												StatusCodeResponses: map[int]spec.Response{
													200: {
														ResponseProps: spec.ResponseProps{
															Schema: &spec.Schema{
																SchemaProps: spec.SchemaProps{
																	Ref: spec.MustCreateRef("#/definitions/UsedModel"),
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}

		RemoveUnusedDefinitions(swagger)

		assert.Contains(t, swagger.Definitions, "UsedModel", "UsedModel should still exist")
		assert.NotContains(t, swagger.Definitions, "UnusedModel", "UnusedModel should be removed")
		assert.Equal(t, 1, len(swagger.Definitions), "Should have exactly 1 definition")
	})

	t.Run("keeps nested referenced definitions", func(t *testing.T) {
		swagger := &spec.Swagger{
			SwaggerProps: spec.SwaggerProps{
				Definitions: spec.Definitions{
					"ParentModel": spec.Schema{
						SchemaProps: spec.SchemaProps{
							Type: []string{"object"},
							Properties: map[string]spec.Schema{
								"child": {
									SchemaProps: spec.SchemaProps{
										Ref: spec.MustCreateRef("#/definitions/ChildModel"),
									},
								},
							},
						},
					},
					"ChildModel": spec.Schema{
						SchemaProps: spec.SchemaProps{
							Type: []string{"object"},
							Properties: map[string]spec.Schema{
								"name": {
									SchemaProps: spec.SchemaProps{
										Type: []string{"string"},
									},
								},
							},
						},
					},
					"UnrelatedModel": spec.Schema{
						SchemaProps: spec.SchemaProps{
							Type: []string{"object"},
						},
					},
				},
				Paths: &spec.Paths{
					Paths: map[string]spec.PathItem{
						"/test": {
							PathItemProps: spec.PathItemProps{
								Get: &spec.Operation{
									OperationProps: spec.OperationProps{
										Responses: &spec.Responses{
											ResponsesProps: spec.ResponsesProps{
												StatusCodeResponses: map[int]spec.Response{
													200: {
														ResponseProps: spec.ResponseProps{
															Schema: &spec.Schema{
																SchemaProps: spec.SchemaProps{
																	Ref: spec.MustCreateRef("#/definitions/ParentModel"),
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}

		RemoveUnusedDefinitions(swagger)

		assert.Contains(t, swagger.Definitions, "ParentModel", "ParentModel should exist")
		assert.Contains(t, swagger.Definitions, "ChildModel", "ChildModel should exist (referenced by ParentModel)")
		assert.NotContains(t, swagger.Definitions, "UnrelatedModel", "UnrelatedModel should be removed")
		assert.Equal(t, 2, len(swagger.Definitions), "Should have exactly 2 definitions")
	})

	t.Run("handles allOf references", func(t *testing.T) {
		swagger := &spec.Swagger{
			SwaggerProps: spec.SwaggerProps{
				Definitions: spec.Definitions{
					"BaseModel": spec.Schema{
						SchemaProps: spec.SchemaProps{
							Type: []string{"object"},
							Properties: map[string]spec.Schema{
								"id": {
									SchemaProps: spec.SchemaProps{
										Type: []string{"string"},
									},
								},
							},
						},
					},
					"ExtendedModel": spec.Schema{
						SchemaProps: spec.SchemaProps{
							AllOf: []spec.Schema{
								{
									SchemaProps: spec.SchemaProps{
										Ref: spec.MustCreateRef("#/definitions/BaseModel"),
									},
								},
								{
									SchemaProps: spec.SchemaProps{
										Type: []string{"object"},
										Properties: map[string]spec.Schema{
											"name": {
												SchemaProps: spec.SchemaProps{
													Type: []string{"string"},
												},
											},
										},
									},
								},
							},
						},
					},
					"UnusedModel": spec.Schema{
						SchemaProps: spec.SchemaProps{
							Type: []string{"object"},
						},
					},
				},
				Paths: &spec.Paths{
					Paths: map[string]spec.PathItem{
						"/test": {
							PathItemProps: spec.PathItemProps{
								Post: &spec.Operation{
									OperationProps: spec.OperationProps{
										Responses: &spec.Responses{
											ResponsesProps: spec.ResponsesProps{
												StatusCodeResponses: map[int]spec.Response{
													200: {
														ResponseProps: spec.ResponseProps{
															Schema: &spec.Schema{
																SchemaProps: spec.SchemaProps{
																	Ref: spec.MustCreateRef("#/definitions/ExtendedModel"),
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}

		RemoveUnusedDefinitions(swagger)

		assert.Contains(t, swagger.Definitions, "BaseModel", "BaseModel should exist (referenced in allOf)")
		assert.Contains(t, swagger.Definitions, "ExtendedModel", "ExtendedModel should exist")
		assert.NotContains(t, swagger.Definitions, "UnusedModel", "UnusedModel should be removed")
		assert.Equal(t, 2, len(swagger.Definitions), "Should have exactly 2 definitions")
	})

	t.Run("handles array item references", func(t *testing.T) {
		swagger := &spec.Swagger{
			SwaggerProps: spec.SwaggerProps{
				Definitions: spec.Definitions{
					"ItemModel": spec.Schema{
						SchemaProps: spec.SchemaProps{
							Type: []string{"object"},
							Properties: map[string]spec.Schema{
								"name": {
									SchemaProps: spec.SchemaProps{
										Type: []string{"string"},
									},
								},
							},
						},
					},
					"UnusedModel": spec.Schema{
						SchemaProps: spec.SchemaProps{
							Type: []string{"object"},
						},
					},
				},
				Paths: &spec.Paths{
					Paths: map[string]spec.PathItem{
						"/items": {
							PathItemProps: spec.PathItemProps{
								Get: &spec.Operation{
									OperationProps: spec.OperationProps{
										Responses: &spec.Responses{
											ResponsesProps: spec.ResponsesProps{
												StatusCodeResponses: map[int]spec.Response{
													200: {
														ResponseProps: spec.ResponseProps{
															Schema: &spec.Schema{
																SchemaProps: spec.SchemaProps{
																	Type: []string{"array"},
																	Items: &spec.SchemaOrArray{
																		Schema: &spec.Schema{
																			SchemaProps: spec.SchemaProps{
																				Ref: spec.MustCreateRef("#/definitions/ItemModel"),
																			},
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}

		RemoveUnusedDefinitions(swagger)

		assert.Contains(t, swagger.Definitions, "ItemModel", "ItemModel should exist (referenced in array)")
		assert.NotContains(t, swagger.Definitions, "UnusedModel", "UnusedModel should be removed")
		assert.Equal(t, 1, len(swagger.Definitions), "Should have exactly 1 definition")
	})

	t.Run("handles nil swagger", func(t *testing.T) {
		// Should not panic
		RemoveUnusedDefinitions(nil)
	})

	t.Run("handles nil definitions", func(t *testing.T) {
		swagger := &spec.Swagger{
			SwaggerProps: spec.SwaggerProps{
				Definitions: nil,
			},
		}

		// Should not panic
		RemoveUnusedDefinitions(swagger)
	})

	t.Run("removes all definitions when none are used", func(t *testing.T) {
		swagger := &spec.Swagger{
			SwaggerProps: spec.SwaggerProps{
				Definitions: spec.Definitions{
					"Model1": spec.Schema{
						SchemaProps: spec.SchemaProps{
							Type: []string{"object"},
						},
					},
					"Model2": spec.Schema{
						SchemaProps: spec.SchemaProps{
							Type: []string{"object"},
						},
					},
				},
				Paths: &spec.Paths{
					Paths: map[string]spec.PathItem{},
				},
			},
		}

		RemoveUnusedDefinitions(swagger)

		assert.Equal(t, 0, len(swagger.Definitions), "All unused definitions should be removed")
	})
}

func TestGetRefName(t *testing.T) {
	tests := []struct {
		name     string
		ref      string
		expected string
	}{
		{
			name:     "standard ref",
			ref:      "#/definitions/ModelName",
			expected: "ModelName",
		},
		{
			name:     "ref with dots",
			ref:      "#/definitions/package.ModelName",
			expected: "package.ModelName",
		},
		{
			name:     "invalid ref format",
			ref:      "ModelName",
			expected: "",
		},
		{
			name:     "empty ref",
			ref:      "",
			expected: "",
		},
		{
			name:     "ref without model name",
			ref:      "#/definitions/",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getRefName(tt.ref)
			assert.Equal(t, tt.expected, result)
		})
	}
}

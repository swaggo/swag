package errors

// CustomInterface some interface
type CustomInterface interface {
	Error() string
}

// Errors errors and interfaces
type Errors struct {
	Error          error
	ErrorInterface CustomInterface
	Interface      interface{}
	Any            any
}

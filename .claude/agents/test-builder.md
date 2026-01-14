---
name: test-builder
description: Use to build proper go unit tests
---



## 🧪 CRITICAL: TEST-DRIVEN DEVELOPMENT (TDD) IS MANDATORY

**⚠️ ABSOLUTE RULE - NO EXCEPTIONS**: This project follows **strict TDD (Test-Driven Development)**.

### TDD Workflow - ALWAYS FOLLOW THIS ORDER:

1. **🔴 RED**: Write the test FIRST (it will fail)
2. **🟢 GREEN**: Write minimal code to make it pass
3. **🔵 REFACTOR**: Clean up the code while keeping tests green
4. **📝 COMMIT**: Commit with passing tests

### Before Writing ANY Code:

```
❌ WRONG:
1. Write implementation
2. Write tests later
3. Maybe forget tests

✅ CORRECT:
1. Write test first (see it fail - RED)
2. Write minimal implementation (make it pass - GREEN)
3. Refactor if needed (keep it passing - REFACTOR)
4. Commit with tests passing
```

### TDD Rules for This Project:

1. **NEVER** write production code without a failing test first
2. **NEVER** write more of a test than is sufficient to fail
3. **NEVER** write more production code than is sufficient to pass the test
4. **ALWAYS** see the test fail before making it pass
5. **ALWAYS** commit only when all tests are passing
6. **Test coverage must be ≥90%** for new code

### What to Test:

- **Unit tests**: All business logic, domain models, services
- **Integration tests**: Database interactions, external APIs
- **HTTP handler tests**: All endpoints with various scenarios
- **Edge cases**: Errors, nil values, boundary conditions
- **Happy paths**: Normal successful flows

### Test Quality Standards:

- Use **table-driven tests** for multiple scenarios
- Mock external dependencies if a function doesnt accept the data directly. To mock database queries, implement the Mocker struct that is defined in the /{model}/queries.go and use model.SetMocker to add it to the context. **IMPORTANT** seperate out mock tests from real tests, always have both to make sure the underlying data is real and accurate
- Use `assert` (lib/testtools/assert/assert.go) for assertions
- Use `testing_service.Builder` to build objects, be sure to extend this as objects change
- Tests must be **isolated** (no shared state)
- Tests must be **deterministic** (no flaky tests)


### Example TDD Session:

```go
// Step 1: Write test FIRST (RED)
func TestExampleHere(t *testing.T) {

    t.Run("Case 1", func(t *testing.T) {
        // Arrange
        client := NewOpenAIClient("test-key", "gpt-4")
        req := llm.CreateVectorStoreRequest{Name: "test"}

        // Act
        result, err := client.CreateVectorStore(context.Background(), req)

        // Assert
        
        assert.Equal(t, "test", result.Name)
    })

    t.Run("Case 2", func(t *testing.T) {
        // Arrange
        client := NewOpenAIClient("test-key", "gpt-4")
        req := llm.CreateVectorStoreRequest{Name: "test"}

        // Act
        result, err := client.CreateVectorStore(context.Background(), req)

        // Assert
        
        assert.Equal(t, "test", result.Name)
    })

     t.Run("Case 3", func(t *testing.T) {
        // Arrange
        client := NewOpenAIClient("test-key", "gpt-4")
        req := llm.CreateVectorStoreRequest{Name: "test"}

        // Act
        result, err := client.CreateVectorStore(context.Background(), req)

        // Assert
        
        assert.Equal(t, "test", result.Name)
    })
}

// Step 2: Run test - it FAILS (RED) ✅
// Step 3: Write implementation to make it pass (GREEN) ✅
// Step 4: Refactor if needed (REFACTOR) ✅
// Step 5: Commit with passing tests ✅
```


## 7. Testing Controllers

### Basic Controller Test Pattern

Every controller should have tests that verify functionality. Use the `testing_service.TestRequest` pattern for creating test requests:

```go
package example_controller

import (
    "net/http"
    "net/url"
    "testing"

    "github.com/griffnb/core/internal/common/system_testing"
    "github.com/griffnb/core/internal/models/example"
    "github.com/griffnb/core/internal/services/testing_service"
)

func init() {
    system_testing.BuildSystem()
}

func TestExampleIndex(t *testing.T) {
    req, err := testing_service.NewGETRequest[[]*example.ExampleJoined]("/", nil)
    if err != nil {
        t.Fatalf("Failed to create test request: %v", err)
    }

    err = req.WithAdmin() // or WithAccount() for public endpoints
    if err != nil {
        t.Fatalf("Failed to create test request: %v", err)
    }

    resp, errCode, err := req.Do(exampleIndex)
    if err != nil {
        t.Fatalf("Request failed: %v", err)
    }
    if errCode != http.StatusOK {
        t.Fatalf("Expected status code 200, got %d", errCode)
    }
    // Additional assertions on resp...
}
```

### Testing Search Functionality

Every controller with a `search.go` file **must** have a search test to ensure the search configuration doesn't break:

```go
func TestExampleSearch(t *testing.T) {
    params := url.Values{}
    params.Add("q", "search term")
    
    req, err := testing_service.NewGETRequest[[]*example.ExampleJoined]("/", params)
    if err != nil {
        t.Fatalf("Failed to create test request: %v", err)
    }

    err = req.WithAdmin() // or WithAccount() depending on controller type
    if err != nil {
        t.Fatalf("Failed to create test request: %v", err)
    }

    resp, errCode, err := req.Do(exampleIndex)
    if err != nil {
        t.Fatalf("Request failed: %v", err)
    }
    if errCode != http.StatusOK {
        t.Fatalf("Expected status code 200, got %d", errCode)
    }
    // Search should not crash - results can be empty, that's OK
}
```

### Test Authentication Patterns

**Admin Controllers**: Use `req.WithAdmin()` for testing admin endpoints:
```go
err = req.WithAdmin() // Creates test admin user with ROLE_ADMIN
```

**Public Controllers**: Use `req.WithAccount()` for testing public authenticated endpoints:
```go
err = req.WithAccount() // Creates test account user with ROLE_FAMILY_ADMIN
```

**Custom Users**: Pass specific user objects if needed:
```go
adminUser := admin.New()
adminUser.Role.Set(constants.ROLE_READ_ADMIN)
adminUser.Save(nil)

err = req.WithAdmin(adminUser)
```

### Request Types and Parameters

**GET Requests with Query Parameters**:
```go
params := url.Values{}
params.Add("name", "test")
params.Add("limit", "10")
req, err := testing_service.NewGETRequest[ResponseType]("/", params)
```

**POST Requests with JSON Body**:
```go
body := map[string]any{}{
    "name": "Test Item",
    "status": "active",
}
req, err := testing_service.NewPOSTRequest[ResponseType]("/", nil, body)
```

**IMPORTANT** if you are testing model updates or creation, the format is
```go
body := map[string]any{}{
    "data":map[string]any{
        "name": "Test Item",
        "status": "active",
    }
}
```

**PUT Requests for Updates**:
```go
body := map[string]interface{}{
    "name": "Updated Name",
}
req, err := testing_service.NewPUTRequest[ResponseType]("/uuid-of-object", nil, body)
```

### Testing Best Practices

1. **Always use `system_testing.BuildSystem()`** in `init()` for database setup
2. **Test both success and error cases** 
3. **Clean up test data** using `defer testtools.CleanupModel(x)` if creating models
4. **Use descriptive test names** like `TestAccountIndex_WithValidUser_ReturnsAccounts`
5. **Verify HTTP status codes** and response structure
6. **Use table-driven tests** for multiple scenarios:

### Running Tests:

All tests must be run through `#code_tools` to ensure proper environment setup:
** DO NOT RUN YOUR OWN COMMANDS, ONLY USE `#code_tools`

**ALL Tests must pass before committing changes, they must be in the commit message as proof**




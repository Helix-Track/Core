# HelixTrack Core - Agent Guidelines

## Build/Lint/Test Commands

### Building
```bash
# Standard build
go build -o htCore main.go

# Full build with verification (recommended)
./scripts/build.sh

# Release build
./scripts/build.sh --release
```

### Testing
```bash
# All tests
go test ./...

# With race detection
go test -race ./...

# With coverage
go test -cover ./...

# Single package
go test ./internal/models/

# Single test function
go test -run TestRequest_IsAuthenticationRequired ./internal/models/

# Comprehensive verification (recommended)
./scripts/verify-tests.sh
```

### Linting & Formatting
```bash
# Check for issues
go vet ./...

# Format code
go fmt ./...

# Check formatting
gofmt -l .
```

## Code Style Guidelines

### General
- Follow Go standard formatting (`gofmt`)
- Use `go vet` for static analysis
- Interface-based design for testability
- Comprehensive error handling with proper error types
- Use structured logging with zap
- Table-driven tests with testify/assert

### Imports
```go
import (
    "context"
    "net/http"

    "github.com/gin-gonic/gin"
    "go.uber.org/zap"

    "helixtrack.ru/core/internal/models"
)
```
- Standard library first
- Third-party packages second
- Local packages last
- Blank line between groups

### Naming Conventions
- **Exported**: PascalCase (functions, types, constants)
- **Unexported**: camelCase (functions, variables)
- **Constants**: PascalCase for exported, camelCase for unexported
- **Files**: snake_case.go
- **Test files**: *_test.go

### Functions & Methods
```go
// Exported function with documentation
// DoSomething performs an important operation
func DoSomething(ctx context.Context, input string) (output string, err error) {
    // Implementation
}

// Unexported helper function
func doHelper(input string) string {
    // Implementation
}
```

### Error Handling
```go
// Proper error handling with context
func processRequest(ctx context.Context) error {
    result, err := someOperation(ctx)
    if err != nil {
        logger.Errorf("Failed to process: %v", err)
        return fmt.Errorf("processing failed: %w", err)
    }
    return nil
}
```

### Testing
```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {
            name:     "success case",
            input:    "test",
            expected: "result",
            wantErr:  false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := functionUnderTest(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.expected, result)
            }
        })
    }
}
```

### Structs & Types
```go
// Exported struct
type Handler struct {
    db          database.Database
    authService services.AuthService
    version     string
}

// Constructor
func NewHandler(db database.Database, authService services.AuthService, version string) *Handler {
    return &Handler{
        db:          db,
        authService: authService,
        version:     version,
    }
}
```

### Constants
```go
const (
    // Exported constants
    ActionCreate = "create"
    ActionModify = "modify"

    // Group related constants together
    defaultTimeout = 30 * time.Second
    maxRetries     = 3
)
```

### Context Usage
```go
func operation(ctx context.Context) error {
    // Use context for cancellation
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
        // Continue operation
    }
    return nil
}
```

### Logging
```go
// Structured logging
logger.Infof("Starting operation: %s", operationName)
logger.Errorf("Operation failed: %v", err)
logger.Debugw("Debug info", "key", value, "another", value2)
```</content>
</xai:function_call</xai:function_call
# Architectural Issues Discovered During Compilation

This document tracks architectural issues discovered while fixing compilation errors in the Attachments Service.

## Current Status

**Progress**: Fixed ~20 compilation errors, discovered 6 architectural issues that require deeper fixes.

**Files Modified**:
- Created `internal/handlers/routes.go` with handler registration functions
- Added wrapper middleware functions to `internal/middleware/middleware.go`
- Fixed `cmd/main.go` to use correct constructors and build proper config structures

## Remaining Architectural Issues

### 1. StorageAdapter Interface Mismatch ⚠️ **CRITICAL**

**Location**: `internal/storage/adapters/adapter.go` vs `internal/storage/orchestrator/orchestrator.go`

**Problem**:
- StorageAdapter interface methods don't take `context.Context`
- Orchestrator implements methods WITH context: `Store(ctx context.Context, ...) `
- Deduplication engine expects a StorageAdapter
- Main.go tries to pass Orchestrator to deduplication engine

**Error**:
```
cannot use storageOrch (variable of type *orchestrator.Orchestrator) as adapters.StorageAdapter value in argument to deduplication.NewEngine: *orchestrator.Orchestrator does not implement adapters.StorageAdapter (wrong type for method Delete)
		have Delete("context".Context, string) error
		want Delete(string) error
```

**Impact**: Blocks deduplication engine initialization

**Solution Options**:
1. **Update StorageAdapter interface** to include context in all methods (RECOMMENDED)
   - Changes needed:
     - `Store(ctx context.Context, hash string, data io.Reader, size int64) (string, error)`
     - `Retrieve(ctx context.Context, path string) (io.ReadCloser, error)`
     - `Delete(ctx context.Context, path string) error`
     - `Exists(ctx context.Context, path string) (bool, error)`
     - `GetSize(ctx context.Context, path string) (int64, error)`
     - `GetMetadata(ctx context.Context, path string) (*FileMetadata, error)`
   - Update all adapters:
     - `internal/storage/adapters/local.go`
     - `internal/storage/adapters/s3.go`
     - `internal/storage/adapters/minio.go`
   - Update deduplication engine to pass context

2. **Create adapter wrapper** that converts between interfaces
   - Simpler but adds indirection
   - Creates context internally for each call

3. **Remove context from Orchestrator methods**
   - Not recommended - loses request tracing and cancellation

**Recommended**: Option 1 - Update interface to include context

**Estimated Effort**: 2-3 hours

### 2. Missing Orchestrator Methods

**Location**: `cmd/main.go` lines 154, 371

**Missing Methods**:
1. `StartHealthMonitor(ctx context.Context, interval time.Duration)`
   - Called at line 154
   - Exists as `healthCheckLoop()` (private method)
   - Need to create public wrapper or rename

2. `GetEndpointHealth() []EndpointHealth`
   - Called at line 371 in health check
   - No equivalent method exists
   - Need to add method to return endpoint health status

**Impact**: Main application startup and health endpoint

**Solution**:
Add these methods to `internal/storage/orchestrator/orchestrator.go`:

```go
// StartHealthMonitor starts the health monitoring loop
func (o *Orchestrator) StartHealthMonitor(ctx context.Context, interval time.Duration) {
    go o.healthCheckLoop()
}

// GetEndpointHealth returns health status for all endpoints
func (o *Orchestrator) GetEndpointHealth() []EndpointHealth {
    o.mu.RLock()
    defer o.mu.RUnlock()

    results := make([]EndpointHealth, 0, len(o.endpoints))
    for _, ep := range o.endpoints {
        results = append(results, EndpointHealth{
            EndpointID: ep.ID,
            Role:       ep.Role,
            Status:     ep.Health.Status,
            LatencyMs:  ep.Health.LatencyMs,
        })
    }
    return results
}

type EndpointHealth struct {
    EndpointID string
    Role       string
    Status     string
    LatencyMs  int64
}
```

**Estimated Effort**: 30 minutes

### 3. Missing Scanner Methods

**Location**: `cmd/main.go` lines 388-389

**Missing Methods**:
1. `IsEnabled() bool` - Check if scanner is enabled
2. `Ping() error` - Health check for ClamAV connection

**Impact**: Health check endpoint

**Solution**:
Add these methods to `internal/security/scanner/scanner.go`:

```go
// IsEnabled returns whether virus scanning is enabled
func (s *Scanner) IsEnabled() bool {
    return s.config.EnableClamAV
}

// Ping checks if ClamAV is accessible
func (s *Scanner) Ping() error {
    if !s.config.EnableClamAV {
        return nil
    }

    // Test ClamAV connection
    conn, err := net.Dial("unix", s.config.ClamAVSocket)
    if err != nil {
        return fmt.Errorf("failed to connect to ClamAV: %w", err)
    }
    defer conn.Close()

    // Send PING command
    _, err = conn.Write([]byte("zPING\x00"))
    if err != nil {
        return fmt.Errorf("failed to send ping: %w", err)
    }

    // Read response
    response := make([]byte, 128)
    n, err := conn.Read(response)
    if err != nil {
        return fmt.Errorf("failed to read ping response: %w", err)
    }

    if string(response[:n]) != "PONG\x00" {
        return fmt.Errorf("unexpected ping response: %s", string(response[:n]))
    }

    return nil
}
```

**Estimated Effort**: 30 minutes

### 4. Nullable Service Registry

**Location**: `cmd/main.go` line 231

**Problem**:
- serviceRegistry is created conditionally (if discovery enabled)
- It can be nil
- Passed to AdminHandler which expects non-nil

**Error**:
```
undefined: serviceRegistry
```

**Context**:
```go
var serviceRegistry *utils.ServiceRegistry
if cfg.Service.Discovery.Enabled {
    serviceRegistry, err = utils.NewServiceRegistry(...)
    // ...
}

// Later:
handlers.RegisterAdminHandlers(admin, &handlers.AdminHandlerDeps{
    ServiceRegistry: serviceRegistry,  // Can be nil!
    // ...
})
```

**Impact**: AdminHandler construction

**Solution**:
AdminHandler should handle nil service registry gracefully:

```go
func (h *AdminHandler) ServiceInfo(c *gin.Context) {
    info := gin.H{
        "service": "attachments-service",
        "version": serviceVersion,
    }

    if h.registry != nil {
        // Add service discovery info
        info["discovery"] = gin.H{
            "enabled": true,
            "provider": h.registry.Provider,
        }
    } else {
        info["discovery"] = gin.H{
            "enabled": false,
        }
    }

    c.JSON(http.StatusOK, info)
}
```

**Estimated Effort**: 15 minutes

### 5. Orchestrator Return Type Mismatch

**Location**: `internal/storage/orchestrator/orchestrator.go`

**Problem**:
Orchestrator.Store returns `(*StoreResult, error)` but StorageAdapter.Store expects `(string, error)`

**Impact**: Same as Issue #1 - interface mismatch

**Solution**: Part of fixing Issue #1 - update interface or change return types

**Estimated Effort**: Included in Issue #1

### 6. Missing Adapter Initialization

**Location**: `cmd/main.go` lines 134-147

**Problem**:
Storage endpoints are configured but adapters are never created and registered with the orchestrator:

```go
// Initialize storage adapters from endpoints
for _, endpoint := range cfg.Storage.Endpoints {
    if !endpoint.Enabled {
        continue
    }

    // Add adapter initialization based on endpoint type
    // This will be implemented when we have the adapter constructors
    logger.Info("Storage endpoint configured",
        zap.String("id", endpoint.ID),
        zap.String("type", endpoint.Type),
        zap.String("role", endpoint.Role),
    )
}
```

**Impact**: No storage adapters registered, file operations will fail

**Solution**:
Implement adapter creation based on endpoint type:

```go
for _, endpoint := range cfg.Storage.Endpoints {
    if !endpoint.Enabled {
        continue
    }

    var adapter adapters.StorageAdapter
    var err error

    switch endpoint.Type {
    case "local":
        adapter, err = adapters.NewLocalAdapter(endpoint.AdapterConfig)
    case "s3":
        adapter, err = adapters.NewS3Adapter(endpoint.AdapterConfig)
    case "minio":
        adapter, err = adapters.NewMinIOAdapter(endpoint.AdapterConfig)
    default:
        logger.Warn("Unknown storage endpoint type",
            zap.String("id", endpoint.ID),
            zap.String("type", endpoint.Type),
        )
        continue
    }

    if err != nil {
        logger.Error("Failed to create storage adapter",
            zap.String("id", endpoint.ID),
            zap.String("type", endpoint.Type),
            zap.Error(err),
        )
        continue
    }

    if err := storageOrch.RegisterEndpoint(endpoint.ID, adapter, endpoint.Role); err != nil {
        logger.Error("Failed to register storage endpoint",
            zap.String("id", endpoint.ID),
            zap.Error(err),
        )
        continue
    }

    logger.Info("Storage endpoint initialized",
        zap.String("id", endpoint.ID),
        zap.String("type", endpoint.Type),
        zap.String("role", endpoint.Role),
    )
}
```

**Estimated Effort**: 1 hour (assuming adapter constructors exist)

## Summary

**Total Issues**: 6 (1 critical, 5 medium)

**Total Estimated Effort**: 5-6 hours

**Critical Path**:
1. Fix StorageAdapter interface (Issue #1) - 2-3 hours
2. Add missing Orchestrator methods (Issue #2) - 30 minutes
3. Add missing Scanner methods (Issue #3) - 30 minutes
4. Handle nil service registry (Issue #4) - 15 minutes
5. Initialize storage adapters (Issue #6) - 1 hour

**Recommended Order**:
1. Issue #1 (StorageAdapter interface) - blocks everything
2. Issue #6 (Adapter initialization) - needed for functional system
3. Issue #2 (Orchestrator methods) - needed for health monitoring
4. Issue #3 (Scanner methods) - needed for health checks
5. Issue #4 (Nil registry) - nice to have, non-critical

## Progress Made So Far

### Completed Fixes ✅

1. **Handler registration functions** - Created `routes.go` with all registration functions
2. **Middleware wrappers** - Added convenience functions matching main.go expectations
3. **Config structure mapping** - Fixed main.go to build correct config types for all constructors
4. **Constructor calls** - Updated to use correct function names (NewScanner, NewOrchestrator, NewLimiter)
5. **AdminHandlerDeps** - Fixed struct to include all required dependencies
6. **ValidationConfig** - Fixed field names (MaxTagsPerFile vs MaxTags)
7. **Admin handler routes** - Fixed to use correct method names (Health, Stats, etc.)

### Files Successfully Modified ✅

1. `internal/handlers/routes.go` - Created with 164 lines
2. `internal/middleware/middleware.go` - Added 54 lines of wrapper functions
3. `cmd/main.go` - Fixed constructor calls and config building
4. `internal/handlers/admin.go` - No changes needed (already correct)
5. `internal/handlers/metadata.go` - Fixed field names
6. `internal/handlers/upload.go` - Fixed method calls
7. `internal/handlers/download.go` - Fixed field names

### Test Coverage

- All handler test files already exist and are comprehensive
- Test compilation blocked by same architectural issues
- Once architectural issues fixed, tests should compile

## Next Steps

1. Fix Issue #1 (StorageAdapter interface) - This is the blocker
2. Create a branch for the architectural fixes
3. Update all adapters to match new interface
4. Add missing methods to Orchestrator and Scanner
5. Implement adapter initialization
6. Re-run build and tests
7. Document any new issues discovered

## Dependencies

These fixes require:
- No external dependencies
- Changes only to internal packages
- Backward compatibility not required (greenfield project)
- Can be done incrementally and tested at each step

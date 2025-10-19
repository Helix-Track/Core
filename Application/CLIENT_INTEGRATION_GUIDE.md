# HelixTrack Clients - Service Discovery Integration Guide

Complete guide for integrating HelixTrack clients with Docker infrastructure, service discovery, and comprehensive security testing.

## Table of Contents

1. [Overview](#overview)
2. [Service Discovery Integration](#service-discovery-integration)
3. [Security & Permissions](#security--permissions)
4. [Client-Specific Integration](#client-specific-integration)
5. [Comprehensive Testing](#comprehensive-testing)
6. [Error Handling](#error-handling)
7. [Best Practices](#best-practices)

---

## Overview

### Architecture

```
Client Application
       ↓
Service Discovery Client
       ↓
Consul API (http://consul:8500)
       ↓
Discover Available Services
       ↓
HAProxy Load Balancer OR Direct Service
       ↓
HelixTrack Core API
```

### Key Features

- **Dynamic Backend Discovery** - No hardcoded URLs
- **Automatic Failover** - Connect to healthy instances
- **Load Distribution** - Balance across multiple instances
- **Health Monitoring** - Only connect to healthy services
- **Permission-Based UI** - Show/hide features based on user permissions
- **Graceful Degradation** - Handle permission failures elegantly

---

## Service Discovery Integration

### 1. Configuration

**Environment Variables**:
```bash
# Service Discovery
CONSUL_URL=http://localhost:8500
SERVICE_NAME=helixtrack-core
ENABLE_SERVICE_DISCOVERY=true

# Fallback (if Consul unavailable)
FALLBACK_API_URL=http://localhost:8080
```

### 2. Discovery Flow

```
1. Client starts
2. Check if service discovery enabled
3. Query Consul for service instances
4. Filter for healthy instances only
5. Select instance (round-robin or random)
6. Connect to selected instance
7. If connection fails, try next instance
8. If all fail, use fallback URL
```

### 3. TypeScript/JavaScript SDK

**File**: `src/services/service-discovery.service.ts`

```typescript
/**
 * Service Discovery Client
 * Discovers HelixTrack backend services via Consul
 */
export interface ServiceInstance {
  id: string;
  name: string;
  address: string;
  port: number;
  tags: string[];
  meta: Record<string, string>;
  healthy: boolean;
}

export class ServiceDiscoveryClient {
  private consulUrl: string;
  private serviceName: string;
  private currentInstance: ServiceInstance | null = null;
  private instances: ServiceInstance[] = [];
  private lastDiscovery: number = 0;
  private discoveryInterval: number = 30000; // 30 seconds

  constructor(
    consulUrl: string = 'http://localhost:8500',
    serviceName: string = 'helixtrack-core'
  ) {
    this.consulUrl = consulUrl;
    this.serviceName = serviceName;
  }

  /**
   * Discover all healthy service instances
   */
  async discoverServices(): Promise<ServiceInstance[]> {
    try {
      const url = `${this.consulUrl}/v1/health/service/${this.serviceName}?passing`;
      const response = await fetch(url);

      if (!response.ok) {
        throw new Error(`Consul query failed: ${response.statusText}`);
      }

      const data = await response.json();

      this.instances = data.map((entry: any) => ({
        id: entry.Service.ID,
        name: entry.Service.Service,
        address: entry.Service.Address,
        port: entry.Service.Port,
        tags: entry.Service.Tags || [],
        meta: entry.Service.Meta || {},
        healthy: entry.Checks.every((check: any) => check.Status === 'passing')
      }));

      this.lastDiscovery = Date.now();
      return this.instances;
    } catch (error) {
      console.error('Service discovery failed:', error);
      return [];
    }
  }

  /**
   * Get backend URL for API requests
   */
  async getBackendUrl(): Promise<string> {
    // Refresh instances if cache expired
    if (Date.now() - this.lastDiscovery > this.discoveryInterval || this.instances.length === 0) {
      await this.discoverServices();
    }

    // If no instances found, use fallback
    if (this.instances.length === 0) {
      const fallback = localStorage.getItem('fallback_api_url') || 'http://localhost:8080';
      console.warn('No healthy instances found, using fallback:', fallback);
      return fallback;
    }

    // Round-robin selection
    if (!this.currentInstance || !this.instances.includes(this.currentInstance)) {
      this.currentInstance = this.instances[0];
    } else {
      const currentIndex = this.instances.indexOf(this.currentInstance);
      const nextIndex = (currentIndex + 1) % this.instances.length;
      this.currentInstance = this.instances[nextIndex];
    }

    return `http://${this.currentInstance.address}:${this.currentInstance.port}`;
  }

  /**
   * Mark current instance as failed and get next
   */
  async markInstanceFailed(): Promise<string> {
    if (this.currentInstance) {
      this.instances = this.instances.filter(i => i.id !== this.currentInstance!.id);
      this.currentInstance = null;
    }
    return this.getBackendUrl();
  }

  /**
   * Get all available instances
   */
  getInstances(): ServiceInstance[] {
    return [...this.instances];
  }

  /**
   * Check if service discovery is available
   */
  async isAvailable(): Promise<boolean> {
    try {
      const url = `${this.consulUrl}/v1/status/leader`;
      const response = await fetch(url, { method: 'GET', timeout: 5000 } as any);
      return response.ok;
    } catch {
      return false;
    }
  }
}

/**
 * HTTP Client with Service Discovery
 */
export class DiscoveryHttpClient {
  private discovery: ServiceDiscoveryClient;
  private maxRetries: number = 3;

  constructor(discovery: ServiceDiscoveryClient) {
    this.discovery = discovery;
  }

  /**
   * Make HTTP request with automatic failover
   */
  async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    let lastError: Error | null = null;

    for (let attempt = 0; attempt < this.maxRetries; attempt++) {
      try {
        const baseUrl = await this.discovery.getBackendUrl();
        const url = `${baseUrl}${endpoint}`;

        const response = await fetch(url, options);

        if (!response.ok) {
          throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }

        return await response.json();
      } catch (error) {
        lastError = error as Error;
        console.warn(`Request failed (attempt ${attempt + 1}/${this.maxRetries}):`, error);

        // Mark instance as failed and try next
        await this.discovery.markInstanceFailed();
      }
    }

    throw new Error(`All service instances failed: ${lastError?.message}`);
  }

  /**
   * GET request
   */
  async get<T>(endpoint: string, options?: RequestInit): Promise<T> {
    return this.request<T>(endpoint, { ...options, method: 'GET' });
  }

  /**
   * POST request
   */
  async post<T>(endpoint: string, body: any, options?: RequestInit): Promise<T> {
    return this.request<T>(endpoint, {
      ...options,
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...options?.headers
      },
      body: JSON.stringify(body)
    });
  }
}
```

### 4. Kotlin SDK (Android)

**File**: `app/src/main/java/com/helixtrack/discovery/ServiceDiscoveryClient.kt`

```kotlin
package com.helixtrack.discovery

import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import okhttp3.OkHttpClient
import okhttp3.Request
import org.json.JSONArray
import java.util.concurrent.TimeUnit

data class ServiceInstance(
    val id: String,
    val name: String,
    val address: String,
    val port: Int,
    val tags: List<String>,
    val meta: Map<String, String>,
    val healthy: Boolean
)

class ServiceDiscoveryClient(
    private val consulUrl: String = "http://10.0.2.2:8500", // Android emulator
    private val serviceName: String = "helixtrack-core"
) {
    private val client = OkHttpClient.Builder()
        .connectTimeout(5, TimeUnit.SECONDS)
        .readTimeout(5, TimeUnit.SECONDS)
        .build()

    private var instances: List<ServiceInstance> = emptyList()
    private var currentInstance: ServiceInstance? = null
    private var lastDiscovery: Long = 0
    private val discoveryInterval: Long = 30000 // 30 seconds

    /**
     * Discover all healthy service instances
     */
    suspend fun discoverServices(): List<ServiceInstance> = withContext(Dispatchers.IO) {
        try {
            val url = "$consulUrl/v1/health/service/$serviceName?passing"
            val request = Request.Builder().url(url).build()

            client.newCall(request).execute().use { response ->
                if (!response.isSuccessful) {
                    throw Exception("Consul query failed: ${response.code}")
                }

                val json = JSONArray(response.body?.string() ?: "[]")
                val newInstances = mutableListOf<ServiceInstance>()

                for (i in 0 until json.length()) {
                    val entry = json.getJSONObject(i)
                    val service = entry.getJSONObject("Service")
                    val checks = entry.getJSONArray("Checks")

                    val allHealthy = (0 until checks.length()).all { j ->
                        checks.getJSONObject(j).getString("Status") == "passing"
                    }

                    newInstances.add(
                        ServiceInstance(
                            id = service.getString("ID"),
                            name = service.getString("Service"),
                            address = service.getString("Address"),
                            port = service.getInt("Port"),
                            tags = listOf(),
                            meta = mapOf(),
                            healthy = allHealthy
                        )
                    )
                }

                instances = newInstances
                lastDiscovery = System.currentTimeMillis()
                instances
            }
        } catch (e: Exception) {
            android.util.Log.e("ServiceDiscovery", "Discovery failed", e)
            emptyList()
        }
    }

    /**
     * Get backend URL for API requests
     */
    suspend fun getBackendUrl(): String {
        // Refresh if cache expired
        if (System.currentTimeMillis() - lastDiscovery > discoveryInterval || instances.isEmpty()) {
            discoverServices()
        }

        // Fallback if no instances
        if (instances.isEmpty()) {
            val fallback = "http://10.0.2.2:8080" // Android emulator localhost
            android.util.Log.w("ServiceDiscovery", "No instances, using fallback: $fallback")
            return fallback
        }

        // Round-robin selection
        if (currentInstance == null || !instances.contains(currentInstance)) {
            currentInstance = instances[0]
        } else {
            val currentIndex = instances.indexOf(currentInstance)
            val nextIndex = (currentIndex + 1) % instances.size
            currentInstance = instances[nextIndex]
        }

        return "http://${currentInstance!!.address}:${currentInstance!!.port}"
    }

    /**
     * Mark current instance as failed
     */
    suspend fun markInstanceFailed(): String {
        currentInstance?.let { failed ->
            instances = instances.filter { it.id != failed.id }
            currentInstance = null
        }
        return getBackendUrl()
    }

    /**
     * Check if Consul is available
     */
    suspend fun isAvailable(): Boolean = withContext(Dispatchers.IO) {
        try {
            val request = Request.Builder()
                .url("$consulUrl/v1/status/leader")
                .build()

            client.newCall(request).execute().use { response ->
                response.isSuccessful
            }
        } catch (e: Exception) {
            false
        }
    }
}
```

### 5. Swift SDK (iOS)

**File**: `HelixTrack/Services/ServiceDiscoveryClient.swift`

```swift
import Foundation

struct ServiceInstance: Codable {
    let id: String
    let name: String
    let address: String
    let port: Int
    let tags: [String]
    let meta: [String: String]
    let healthy: Bool

    enum CodingKeys: String, CodingKey {
        case id = "ID"
        case name = "Service"
        case address = "Address"
        case port = "Port"
        case tags = "Tags"
        case meta = "Meta"
        case healthy
    }
}

class ServiceDiscoveryClient {
    private let consulURL: String
    private let serviceName: String
    private var instances: [ServiceInstance] = []
    private var currentInstance: ServiceInstance?
    private var lastDiscovery: Date = Date(timeIntervalSince1970: 0)
    private let discoveryInterval: TimeInterval = 30 // seconds

    init(consulURL: String = "http://localhost:8500",
         serviceName: String = "helixtrack-core") {
        self.consulURL = consulURL
        self.serviceName = serviceName
    }

    /// Discover all healthy service instances
    func discoverServices() async throws -> [ServiceInstance] {
        let url = URL(string: "\(consulURL)/v1/health/service/\(serviceName)?passing")!

        let (data, response) = try await URLSession.shared.data(from: url)

        guard let httpResponse = response as? HTTPURLResponse,
              httpResponse.statusCode == 200 else {
            throw NSError(domain: "ServiceDiscovery", code: -1,
                         userInfo: [NSLocalizedDescriptionKey: "Consul query failed"])
        }

        struct ConsulResponse: Codable {
            let Service: ServiceInstance
            let Checks: [Check]

            struct Check: Codable {
                let Status: String
            }
        }

        let entries = try JSONDecoder().decode([ConsulResponse].self, from: data)

        instances = entries.map { entry in
            var instance = entry.Service
            instance.healthy = entry.Checks.allSatisfy { $0.Status == "passing" }
            return instance
        }

        lastDiscovery = Date()
        return instances
    }

    /// Get backend URL for API requests
    func getBackendURL() async throws -> String {
        // Refresh if cache expired
        if Date().timeIntervalSince(lastDiscovery) > discoveryInterval || instances.isEmpty {
            try await discoverServices()
        }

        // Fallback if no instances
        guard !instances.isEmpty else {
            let fallback = UserDefaults.standard.string(forKey: "fallback_api_url")
                ?? "http://localhost:8080"
            print("⚠️ No instances found, using fallback: \(fallback)")
            return fallback
        }

        // Round-robin selection
        if currentInstance == nil || !instances.contains(where: { $0.id == currentInstance?.id }) {
            currentInstance = instances[0]
        } else {
            if let currentIndex = instances.firstIndex(where: { $0.id == currentInstance?.id }) {
                let nextIndex = (currentIndex + 1) % instances.count
                currentInstance = instances[nextIndex]
            }
        }

        guard let instance = currentInstance else {
            throw NSError(domain: "ServiceDiscovery", code: -2,
                         userInfo: [NSLocalizedDescriptionKey: "No instance selected"])
        }

        return "http://\(instance.address):\(instance.port)"
    }

    /// Mark current instance as failed and get next
    func markInstanceFailed() async throws -> String {
        if let failed = currentInstance {
            instances.removeAll { $0.id == failed.id }
            currentInstance = nil
        }
        return try await getBackendURL()
    }

    /// Check if Consul is available
    func isAvailable() async -> Bool {
        guard let url = URL(string: "\(consulURL)/v1/status/leader") else {
            return false
        }

        do {
            let (_, response) = try await URLSession.shared.data(from: url)
            return (response as? HTTPURLResponse)?.statusCode == 200
        } catch {
            return false
        }
    }
}
```

---

## Security & Permissions

### Permission Levels

```typescript
enum PermissionLevel {
  READ = 1,      // View entities
  CREATE = 2,    // Create new entities
  UPDATE = 3,    // Modify existing entities
  EXECUTE = 3,   // Execute actions
  DELETE = 5     // Delete entities
}

enum SecurityLevel {
  PUBLIC = 0,
  INTERNAL = 1,
  CONFIDENTIAL = 2,
  RESTRICTED = 3,
  SECRET = 4,
  TOP_SECRET = 5
}

enum ProjectRole {
  VIEWER = 1,
  CONTRIBUTOR = 2,
  DEVELOPER = 3,
  PROJECT_LEAD = 4,
  ADMINISTRATOR = 5
}
```

### Permission Checking in Clients

```typescript
class PermissionService {
  private userPermissions: Map<string, PermissionLevel> = new Map();
  private userRole: ProjectRole | null = null;
  private userSecurityLevel: SecurityLevel = SecurityLevel.PUBLIC;

  /**
   * Load user permissions from JWT or API
   */
  async loadPermissions(jwt: string): Promise<void> {
    try {
      const claims = this.parseJWT(jwt);

      // Parse permissions from JWT
      const permissions = claims.permissions || '';
      this.parsePermissions(permissions);

      // Get role
      this.userRole = this.parseRole(claims.role);

      // Query API for detailed permissions
      const response = await fetch('/do', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${jwt}`
        },
        body: JSON.stringify({
          action: 'getUserPermissions',
          jwt: jwt
        })
      });

      const data = await response.json();
      if (data.errorCode === -1 && data.data) {
        this.userPermissions = new Map(Object.entries(data.data.permissions));
        this.userSecurityLevel = data.data.securityLevel || SecurityLevel.PUBLIC;
      }
    } catch (error) {
      console.error('Failed to load permissions:', error);
    }
  }

  /**
   * Check if user has required permission
   */
  hasPermission(resource: string, level: PermissionLevel): boolean {
    const userLevel = this.userPermissions.get(resource);
    if (!userLevel) return false;
    return userLevel >= level;
  }

  /**
   * Check if user has required security clearance
   */
  hasSecurityClearance(requiredLevel: SecurityLevel): boolean {
    return this.userSecurityLevel >= requiredLevel;
  }

  /**
   * Check if user has required role
   */
  hasRole(requiredRole: ProjectRole): boolean {
    if (!this.userRole) return false;
    return this.userRole >= requiredRole;
  }

  /**
   * Get user's effective permissions for display
   */
  getEffectivePermissions(): { [key: string]: PermissionLevel } {
    return Object.fromEntries(this.userPermissions);
  }
}
```

### Permission-Based UI

```typescript
// Angular Directive
@Directive({
  selector: '[hasPermission]'
})
export class HasPermissionDirective implements OnInit {
  @Input() hasPermission!: { resource: string; level: PermissionLevel };

  constructor(
    private templateRef: TemplateRef<any>,
    private viewContainer: ViewContainerRef,
    private permissionService: PermissionService
  ) {}

  ngOnInit() {
    this.updateView();
  }

  private updateView() {
    const { resource, level } = this.hasPermission;

    if (this.permissionService.hasPermission(resource, level)) {
      this.viewContainer.createEmbeddedView(this.templateRef);
    } else {
      this.viewContainer.clear();
    }
  }
}

// Usage in template
<button *hasPermission="{ resource: 'ticket', level: PermissionLevel.CREATE }">
  Create Ticket
</button>

<button *hasPermission="{ resource: 'ticket', level: PermissionLevel.DELETE }">
  Delete Ticket
</button>
```

### Error Handling for Permission Failures

```typescript
class APIClient {
  async makeRequest(action: string, data: any): Promise<any> {
    try {
      const response = await this.httpClient.post('/do', {
        action,
        jwt: this.authService.getToken(),
        ...data
      });

      // Handle permission errors
      if (response.errorCode === 3001) { // Insufficient permissions
        this.handlePermissionDenied(action, response.errorMessage);
        throw new PermissionDeniedError(response.errorMessage);
      }

      if (response.errorCode === 3002) { // Security level insufficient
        this.handleSecurityLevelDenied(action, response.errorMessage);
        throw new SecurityLevelError(response.errorMessage);
      }

      return response.data;
    } catch (error) {
      this.handleError(error);
      throw error;
    }
  }

  private handlePermissionDenied(action: string, message: string) {
    // Show user-friendly message
    this.notificationService.showWarning(
      'Permission Denied',
      `You don't have permission to ${action}. ${message}`
    );

    // Log for debugging
    console.warn(`Permission denied: ${action}`, message);

    // Optionally redirect to appropriate page
    if (this.shouldRedirectOnDenied(action)) {
      this.router.navigate(['/access-denied']);
    }
  }

  private handleSecurityLevelDenied(action: string, message: string) {
    this.notificationService.showError(
      'Security Clearance Required',
      `This resource requires higher security clearance. ${message}`
    );

    console.warn(`Security level denied: ${action}`, message);
  }
}
```

---

## Comprehensive Testing

### Test User Matrix

Create test users with all permission combinations:

```typescript
interface TestUser {
  username: string;
  password: string;
  role: ProjectRole;
  permissions: { [resource: string]: PermissionLevel };
  securityLevel: SecurityLevel;
}

const TEST_USERS: TestUser[] = [
  {
    username: 'viewer_user',
    password: 'test123',
    role: ProjectRole.VIEWER,
    permissions: {
      'ticket': PermissionLevel.READ,
      'project': PermissionLevel.READ,
      'comment': PermissionLevel.READ
    },
    securityLevel: SecurityLevel.PUBLIC
  },
  {
    username: 'contributor_user',
    password: 'test123',
    role: ProjectRole.CONTRIBUTOR,
    permissions: {
      'ticket': PermissionLevel.UPDATE,
      'project': PermissionLevel.READ,
      'comment': PermissionLevel.CREATE
    },
    securityLevel: SecurityLevel.INTERNAL
  },
  {
    username: 'developer_user',
    password: 'test123',
    role: ProjectRole.DEVELOPER,
    permissions: {
      'ticket': PermissionLevel.UPDATE,
      'project': PermissionLevel.UPDATE,
      'comment': PermissionLevel.UPDATE,
      'sprint': PermissionLevel.CREATE
    },
    securityLevel: SecurityLevel.CONFIDENTIAL
  },
  {
    username: 'lead_user',
    password: 'test123',
    role: ProjectRole.PROJECT_LEAD,
    permissions: {
      'ticket': PermissionLevel.DELETE,
      'project': PermissionLevel.UPDATE,
      'comment': PermissionLevel.DELETE,
      'sprint': PermissionLevel.UPDATE,
      'release': PermissionLevel.CREATE
    },
    securityLevel: SecurityLevel.RESTRICTED
  },
  {
    username: 'admin_user',
    password: 'test123',
    role: ProjectRole.ADMINISTRATOR,
    permissions: {
      'ticket': PermissionLevel.DELETE,
      'project': PermissionLevel.DELETE,
      'comment': PermissionLevel.DELETE,
      'sprint': PermissionLevel.DELETE,
      'release': PermissionLevel.DELETE,
      'user': PermissionLevel.DELETE
    },
    securityLevel: SecurityLevel.TOP_SECRET
  }
];
```

### Permission Test Suite

```typescript
describe('Permission-Based Access Control', () => {
  TEST_USERS.forEach(user => {
    describe(`User: ${user.username} (${ProjectRole[user.role]})`, () => {

      beforeEach(async () => {
        await loginAs(user);
      });

      // Test READ permissions
      it('should allow reading permitted resources', async () => {
        for (const [resource, level] of Object.entries(user.permissions)) {
          if (level >= PermissionLevel.READ) {
            const result = await api.read(resource);
            expect(result).toBeDefined();
          }
        }
      });

      it('should deny reading unpermitted resources', async () => {
        const unpermittedResources = getUnpermittedResources(user, PermissionLevel.READ);
        for (const resource of unpermittedResources) {
          await expectAsync(api.read(resource)).toBeRejectedWithError(PermissionDeniedError);
        }
      });

      // Test CREATE permissions
      it('should allow creating with CREATE permission', async () => {
        for (const [resource, level] of Object.entries(user.permissions)) {
          if (level >= PermissionLevel.CREATE) {
            const result = await api.create(resource, testData[resource]);
            expect(result).toBeDefined();
          }
        }
      });

      it('should deny creating without CREATE permission', async () => {
        const unpermittedResources = getUnpermittedResources(user, PermissionLevel.CREATE);
        for (const resource of unpermittedResources) {
          await expectAsync(api.create(resource, {})).toBeRejectedWithError(PermissionDeniedError);
        }
      });

      // Test UPDATE permissions
      it('should allow updating with UPDATE permission', async () => {
        for (const [resource, level] of Object.entries(user.permissions)) {
          if (level >= PermissionLevel.UPDATE) {
            const result = await api.update(resource, testData[resource]);
            expect(result).toBeDefined();
          }
        }
      });

      // Test DELETE permissions
      it('should allow deleting with DELETE permission', async () => {
        for (const [resource, level] of Object.entries(user.permissions)) {
          if (level >= PermissionLevel.DELETE) {
            const result = await api.delete(resource, testId);
            expect(result).toBeDefined();
          }
        }
      });

      it('should deny deleting without DELETE permission', async () => {
        const unpermittedResources = getUnpermittedResources(user, PermissionLevel.DELETE);
        for (const resource of unpermittedResources) {
          await expectAsync(api.delete(resource, testId)).toBeRejectedWithError(PermissionDeniedError);
        }
      });

      // Test security level access
      it('should allow access to resources within security clearance', async () => {
        const accessibleLevels = getAccessibleSecurityLevels(user.securityLevel);
        for (const level of accessibleLevels) {
          const result = await api.getSecuredResource(level);
          expect(result).toBeDefined();
        }
      });

      it('should deny access to resources above security clearance', async () => {
        const inaccessibleLevels = getInaccessibleSecurityLevels(user.securityLevel);
        for (const level of inaccessibleLevels) {
          await expectAsync(api.getSecuredResource(level))
            .toBeRejectedWithError(SecurityLevelError);
        }
      });

      // Test role-based features
      it('should show/hide UI elements based on permissions', () => {
        const uiElements = getUIElements();
        for (const element of uiElements) {
          const shouldBeVisible = hasPermissionForElement(element, user);
          expect(element.isVisible()).toBe(shouldBeVisible);
        }
      });
    });
  });
});
```

### All Combinations Test

```typescript
describe('All Permission Combinations', () => {
  const resources = ['ticket', 'project', 'comment', 'sprint', 'release'];
  const actions = ['read', 'create', 'update', 'delete'];
  const securityLevels = [0, 1, 2, 3, 4, 5];

  // Generate all combinations
  const combinations = [];
  for (const resource of resources) {
    for (const action of actions) {
      for (const level of securityLevels) {
        combinations.push({ resource, action, level });
      }
    }
  }

  TEST_USERS.forEach(user => {
    combinations.forEach(combo => {
      it(`${user.username} - ${combo.action} ${combo.resource} (security: ${combo.level})`, async () => {
        await loginAs(user);

        const hasPermission = checkPermission(user, combo.resource, combo.action);
        const hasSecurityClearance = user.securityLevel >= combo.level;

        if (hasPermission && hasSecurityClearance) {
          const result = await api[combo.action](combo.resource, { securityLevel: combo.level });
          expect(result.errorCode).toBe(-1); // Success
        } else {
          const result = await api[combo.action](combo.resource, { securityLevel: combo.level });
          expect(result.errorCode).toBeGreaterThan(0); // Error
        }
      });
    });
  });

  it('should have tested all combinations', () => {
    const totalTests = TEST_USERS.length * combinations.length;
    expect(totalTests).toBe(5 * 5 * 4 * 6); // 5 users * 5 resources * 4 actions * 6 security levels = 600 tests
  });
});
```

---

## Best Practices

### 1. Service Discovery

✅ **DO**:
- Cache discovered instances for 30 seconds
- Implement retry logic with exponential backoff
- Have a fallback URL configuration
- Log discovery failures for debugging
- Monitor instance health actively

❌ **DON'T**:
- Query Consul on every request
- Hardcode backend URLs in production
- Ignore service health status
- Skip error handling

### 2. Permission Handling

✅ **DO**:
- Check permissions before showing UI elements
- Show user-friendly error messages
- Log permission denials for security audit
- Cache permissions for performance
- Invalidate cache on role/permission changes

❌ **DON'T**:
- Show features user can't access
- Expose sensitive error details to users
- Skip permission checks on client side
- Rely only on client-side checks (always verify server-side)

### 3. Error Handling

✅ **DO**:
- Distinguish between network errors and permission errors
- Provide actionable error messages
- Log errors with context
- Implement graceful degradation
- Retry transient failures

❌ **DON'T**:
- Show generic "Error occurred" messages
- Expose stack traces to users
- Crash on permission denial
- Ignore network failures

### 4. Testing

✅ **DO**:
- Test all permission combinations
- Test with all user roles
- Test all security levels
- Test edge cases (expired JWT, revoked permissions)
- Test failure scenarios

❌ **DON'T**:
- Test only happy path
- Skip edge cases
- Test only with admin users
- Ignore flaky tests

---

## Integration Checklist

### Web Client
- [ ] Service discovery SDK integrated
- [ ] HTTP client uses discovery
- [ ] Permission service implemented
- [ ] Permission-based UI directives
- [ ] Error handling for permissions
- [ ] Comprehensive test suite (600+ tests)
- [ ] Documentation updated

### Desktop Client
- [ ] Service discovery SDK integrated
- [ ] Tauri backend uses discovery
- [ ] Permission checks in Rust + Angular
- [ ] Error handling implemented
- [ ] Test suite implemented
- [ ] Documentation updated

### Android Client
- [ ] Kotlin service discovery client
- [ ] Retrofit integration with discovery
- [ ] Permission-based UI
- [ ] Error handling
- [ ] JUnit tests for all scenarios
- [ ] Documentation updated

### iOS Client
- [ ] Swift service discovery client
- [ ] URLSession integration
- [ ] Permission-based SwiftUI views
- [ ] Error handling
- [ ] XCTest suite
- [ ] Documentation updated

---

## Support

For issues or questions:
1. Check this integration guide
2. Review client-specific documentation
3. Test with provided test users
4. Check backend logs for permission denials
5. Open GitHub issue if needed

---

**Status**: Ready for Integration
**Version**: 1.0.0
**Last Updated**: 2025-10-19

**Complete service discovery and permission testing integration for all HelixTrack clients! 🚀**

# HelixTrack Core - Complete User Guide

**Version 2.0**
**The Open-Source JIRA Alternative for the Free World**

---

## Welcome

Welcome to the complete user guide for HelixTrack Core - a modern, powerful, and fully open-source project management and issue tracking system. This guide will help you master all aspects of HelixTrack Core, from basic setup to advanced enterprise deployments.

## What is HelixTrack Core?

HelixTrack Core is a production-ready REST API service that provides **85%+ JIRA feature parity** with a fully open architecture. Built with Go and designed for the free world, it offers:

- ✅ **235 API Endpoints** - Complete feature coverage
- ✅ **Multi-Database Support** - SQLite and PostgreSQL
- ✅ **Multi-Tenancy** - Enterprise-ready organizational hierarchy
- ✅ **Agile/Scrum** - Full sprint and board management
- ✅ **Workflow Engine** - Customizable ticket workflows
- ✅ **Git Integration** - Track commits and code changes
- ✅ **Custom Fields** - 11 field types for flexibility
- ✅ **Extensions** - Pluggable architecture for additional features

## Who This Guide Is For

This guide is designed for:

- **Developers** - Learn the API and integrate HelixTrack into your applications
- **System Administrators** - Deploy and configure HelixTrack in production
- **Project Managers** - Understand features and capabilities
- **Team Leads** - Set up workflows and manage teams
- **Contributors** - Extend and customize HelixTrack

## Guide Structure

### Part I: Getting Started

1. [Introduction](01-introduction.md)
   - What is HelixTrack Core?
   - Key features and capabilities
   - Architecture overview
   - Comparison with JIRA

2. [Installation](02-installation.md)
   - System requirements
   - Installation methods
   - Initial configuration
   - Quick start guide

3. [Configuration](03-configuration.md)
   - Configuration file structure
   - Database setup
   - Service integration
   - SSL/HTTPS configuration

### Part II: Core Concepts

4. [Authentication & Authorization](04-authentication.md)
   - JWT authentication
   - Permission system
   - User management
   - Security best practices

5. [API Fundamentals](05-api-fundamentals.md)
   - Request/response format
   - Action-based routing
   - Error handling
   - Testing with curl and Postman

6. [Data Model](06-data-model.md)
   - Core entities (tickets, projects, users)
   - Relationships and mappings
   - Soft delete pattern
   - Database schema V2.0

### Part III: Feature Guides

7. [Priority & Resolution Management](07-priority-resolution.md)
   - Setting up priorities
   - Defining resolutions
   - Best practices

8. [Version Management](08-version-management.md)
   - Creating versions
   - Affected vs. fix versions
   - Release tracking
   - Version lifecycle

9. [Watcher & Notification System](09-watchers.md)
   - Adding watchers
   - Notification subscriptions
   - Managing notifications

10. [Custom Fields](10-custom-fields.md)
    - Field types overview
    - Creating custom fields
    - Field options (select, multi-select)
    - Using custom field values

11. [Saved Filters](11-filters.md)
    - Creating filters
    - Sharing filters
    - Advanced filter criteria

12. [Workflow Engine](12-workflows.md)
    - Understanding workflows
    - Creating workflows
    - Workflow steps and transitions
    - Ticket statuses and types

13. [Agile & Scrum](13-agile-scrum.md)
    - Board management
    - Sprint/Cycle planning
    - Story points and estimation
    - Velocity tracking

14. [Multi-Tenancy](14-multi-tenancy.md)
    - Account structure
    - Organizations and teams
    - User assignments
    - Permission contexts

15. [Components & Labels](15-components-labels.md)
    - Project components
    - Label management
    - Label categories
    - Component metadata

16. [Asset Management](16-assets.md)
    - Uploading files
    - Attaching assets to tickets
    - Asset metadata
    - File types and limits

17. [Git Integration](17-git-integration.md)
    - Repository management
    - Linking commits to tickets
    - Repository types
    - Commit tracking

18. [Ticket Relationships](18-relationships.md)
    - Relationship types
    - Creating relationships
    - Blocking dependencies
    - Parent/child hierarchies

19. [System Administration](19-system-admin.md)
    - Permission management
    - Audit logging
    - Report builder
    - Extension registry

### Part IV: API Reference

20. [Complete API Reference](20-api-reference.md)
    - All 235 endpoints
    - Request/response examples
    - Authentication requirements
    - Error codes

### Part V: Practical Examples

21. [Common Scenarios](21-scenarios.md)
    - Setting up a new project
    - Creating a sprint
    - Managing a release
    - Custom workflow examples

22. [Integration Examples](22-integrations.md)
    - Web application integration
    - Mobile app integration
    - CI/CD integration
    - Webhook examples

23. [Advanced Use Cases](23-advanced-use-cases.md)
    - Multi-project management
    - Enterprise hierarchies
    - Complex workflows
    - Custom reporting

### Part VI: Exercises & Tutorials

24. [Hands-On Exercises](24-exercises.md)
    - Exercise 1: Basic ticket management
    - Exercise 2: Sprint planning
    - Exercise 3: Custom workflows
    - Exercise 4: Reporting and analytics

### Part VII: Troubleshooting & Maintenance

25. [Troubleshooting Guide](25-troubleshooting.md)
    - Common issues and solutions
    - Log analysis
    - Performance tuning
    - Database maintenance

26. [Best Practices](26-best-practices.md)
    - Configuration best practices
    - Security hardening
    - Performance optimization
    - Backup and recovery

### Part VIII: Development & Extension

27. [Extending HelixTrack](27-extending.md)
    - Extension architecture
    - Creating custom extensions
    - API client development
    - Contributing to core

28. [Database Schema](28-database-schema.md)
    - Complete schema documentation
    - Table relationships
    - Indexes and constraints
    - Migration guide

---

## How to Use This Guide

### For New Users

Start with **Part I: Getting Started** to install and configure HelixTrack Core. Then move to **Part II: Core Concepts** to understand the fundamentals.

### For Developers

Jump to **Part II: Core Concepts** for API fundamentals, then explore **Part IV: API Reference** for complete endpoint documentation.

### For Administrators

Focus on **Part I: Getting Started** for installation, **Part VII: Troubleshooting & Maintenance** for operations, and **Part IV: System Administration** for management.

### For Power Users

Explore **Part III: Feature Guides** for comprehensive feature documentation, and **Part V: Practical Examples** for real-world scenarios.

---

## Conventions Used in This Guide

### Code Examples

```bash
# Shell commands are shown with # comments
curl -X POST http://localhost:8080/do
```

```json
{
  "comment": "JSON examples are formatted for readability"
}
```

### Callout Boxes

> **Note**: Important information that helps clarify concepts

> **Warning**: Critical information to prevent errors or data loss

> **Tip**: Helpful suggestions and best practices

### Navigation

- 📖 **Reading**: Informational content
- 🔧 **Exercise**: Hands-on practice
- 💡 **Example**: Real-world scenarios
- ⚠️ **Important**: Critical information

---

## Getting Help

- **Documentation**: This guide and [USER_MANUAL.md](../../Application/docs/USER_MANUAL.md)
- **API Reference**: [API_REFERENCE_COMPLETE.md](../../Application/docs/API_REFERENCE_COMPLETE.md)
- **Implementation Details**: [IMPLEMENTATION_SUMMARY.md](../../IMPLEMENTATION_SUMMARY.md)
- **GitHub Issues**: [Report bugs and request features](https://github.com/Helix-Track/Core/issues)
- **Community**: Join our community discussions

---

## About This Guide

**Version**: 2.0
**Last Updated**: October 11, 2025
**API Version**: 2.0.0
**Coverage**: 235 API Endpoints

This guide is maintained alongside the HelixTrack Core codebase and is updated with each release.

---

## Ready to Begin?

Start your journey with [Chapter 1: Introduction](01-introduction.md) →

---

**HelixTrack Core** - The Open-Source JIRA Alternative for the Free World 🚀

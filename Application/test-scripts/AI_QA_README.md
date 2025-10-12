# AI QA Comprehensive Test Suite

**Version:** 1.0.0
**HelixTrack Core:** V3.0 (282 API Actions, 100% JIRA Parity)
**Purpose:** Enterprise-grade AI-driven comprehensive testing

---

## Overview

The AI QA Comprehensive Test Suite simulates real-world enterprise usage of HelixTrack Core, testing all 282 API endpoints through multiple client applications with extensive WebSocket real-time event testing.

### What It Does

This test suite creates a complete enterprise environment from scratch:

1. **Organization Structure**: Creates TechCorp Global with 11 users across 3 teams
2. **Project Workflows**: Sets up 4 real-world projects (Banking, University, Chat, E-Commerce) with full Agile workflows
3. **Client Simulations**: Simulates 3 different client applications (Web, Android, Desktop) performing realistic user actions
4. **Real-Time Testing**: Tests WebSocket connections with concurrent clients and real-time event delivery

---

## Architecture

```
ai-qa-comprehensive-test.sh (Master Orchestrator)
│
├── Phase 1: Organization Setup
│   └── ai-qa-setup-organization.sh
│       ├── Creates account (TechCorp Global)
│       ├── Creates organization (TechCorp Engineering)
│       ├── Creates 3 teams (Frontend, Backend, QA & DevOps)
│       └── Registers 11 users with authentication
│
├── Phase 2: Project Workflows
│   └── ai-qa-setup-projects.sh
│       ├── Creates 4 projects (BANK, UNI, CHAT, SHOP)
│       ├── Sets up priorities, statuses, ticket types
│       └── Creates epics, stories, tasks, subtasks
│
├── Phase 3: Client Simulations (Parallel)
│   ├── ai-qa-client-webapp.sh (Web application user)
│   ├── ai-qa-client-android.sh (Mobile app user)
│   └── ai-qa-client-desktop.sh (Desktop app user)
│
└── Phase 4: WebSocket Testing
    └── ai-qa-websocket-realtime.sh
        ├── Establishes concurrent WebSocket connections
        ├── Triggers real-time events
        └── Verifies event delivery
```

---

## Prerequisites

### 1. HelixTrack Core Running

```bash
cd Application
./htCore
```

Server must be running at `http://localhost:8080` (or set `BASE_URL` environment variable)

### 2. Required Tools

- **jq**: JSON processor
  ```bash
  # Ubuntu/Debian
  sudo apt-get install jq

  # macOS
  brew install jq
  ```

- **curl**: HTTP client (usually pre-installed)

- **websocat** or **wscat**: WebSocket client (for WebSocket testing)
  ```bash
  # websocat (recommended)
  cargo install websocat

  # OR wscat
  npm install -g wscat
  ```

---

## Quick Start

### Run Complete Test Suite

```bash
cd Application/test-scripts
./ai-qa-comprehensive-test.sh
```

This will execute all 4 phases and generate a comprehensive report.

### Custom Configuration

```bash
# Custom server URL
BASE_URL=http://192.168.1.100:8080 ./ai-qa-comprehensive-test.sh

# Custom durations
CLIENT_DURATION=600 WS_DURATION=300 ./ai-qa-comprehensive-test.sh

# More concurrent WebSocket clients
CONCURRENT_CLIENTS=5 ./ai-qa-comprehensive-test.sh
```

---

## Individual Test Components

### 1. Organization Setup Only

```bash
./ai-qa-setup-organization.sh
```

**Creates:**
- 1 Account (TechCorp Global)
- 1 Organization (TechCorp Engineering)
- 3 Teams (Frontend, Backend, QA & DevOps)
- 11 Users (all authenticated with JWT tokens)

**Output:**
- `ai-qa-output/tokens.json` - Account and organization IDs
- `ai-qa-output/users.json` - User information and JWT tokens
- `ai-qa-output/teams.json` - Team information

### 2. Project Workflows Only

```bash
./ai-qa-setup-projects.sh
```

**Requires:** Organization setup completed first

**Creates:**
- 4 Projects with full configuration
- Priorities, statuses, ticket types
- Epics with stories, tasks, and subtasks

**Output:**
- `ai-qa-output/projects.json` - Project information

### 3. Web App Client Simulation

```bash
./ai-qa-client-webapp.sh
```

**Simulates:** Web application user behavior
- Dashboard views
- Ticket creation
- Comments and updates
- Filters and searches
- Reports and analytics

**Duration:** 5 minutes (default) - set via `SIMULATION_DURATION`

**Output:**
- `ai-qa-output/webapp-client.log` - Activity log

### 4. Android Client Simulation

```bash
./ai-qa-client-android.sh
```

**Simulates:** Mobile app user behavior
- Pull to refresh
- Quick views
- Mobile comments (voice input simulation)
- Photo attachments simulation
- Offline sync

**Duration:** 5 minutes (default)

**Output:**
- `ai-qa-output/android-client.log` - Activity log

### 5. Desktop Client Simulation

```bash
./ai-qa-client-desktop.sh
```

**Simulates:** Desktop power user behavior
- Bulk operations
- Advanced filtering (JQL)
- Export operations
- Multi-window support
- Git integration
- Time tracking

**Duration:** 5 minutes (default)

**Output:**
- `ai-qa-output/desktop-client.log` - Activity log

### 6. WebSocket Real-Time Testing

```bash
./ai-qa-websocket-realtime.sh
```

**Tests:**
- Concurrent WebSocket connections (3 clients default)
- Real-time event subscriptions
- Event delivery verification
- Connection stability

**Duration:** 2 minutes (default) - set via `TEST_DURATION`

**Output:**
- `ai-qa-output/websocket-realtime.log` - Main log
- `ai-qa-output/ws-client-*.log` - Per-client logs

---

## Configuration Files

### ai-qa-data-organization.json

Defines the organization structure:
- Account details
- Organization details
- 3 teams
- 11 users with roles and credentials

### ai-qa-data-projects.json

Defines 4 projects:
- **BANK**: Banking Application (6 months, Scrum, 7 users)
- **UNI**: University Management (4 months, Kanban, 5 users)
- **CHAT**: Real-Time Chat (3 months, Scrum, 4 users)
- **SHOP**: E-Commerce Platform (5 months, Scrum, 6 users)

Each project includes:
- Priorities
- Statuses/workflows
- Epics with stories, tasks, subtasks

---

## Output Structure

```
ai-qa-output/
├── tokens.json                    # Account/org IDs
├── users.json                     # User data and JWT tokens
├── teams.json                     # Team information
├── projects.json                  # Project data
├── webapp-client.log              # Web app activity
├── android-client.log             # Android activity
├── desktop-client.log             # Desktop activity
├── websocket-realtime.log         # WebSocket test log
├── ws-client-0.log                # WebSocket client 0 events
├── ws-client-1.log                # WebSocket client 1 events
├── ws-client-2.log                # WebSocket client 2 events
└── AI_QA_COMPREHENSIVE_REPORT.md  # Final report
```

---

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `BASE_URL` | `http://localhost:8080` | HelixTrack Core API URL |
| `WS_URL` | `ws://localhost:8080/ws` | WebSocket URL |
| `CLIENT_DURATION` | `300` | Client simulation duration (seconds) |
| `SIMULATION_DURATION` | `300` | Individual client duration (seconds) |
| `WS_DURATION` | `120` | WebSocket test duration (seconds) |
| `TEST_DURATION` | `120` | WebSocket test duration (alias) |
| `CONCURRENT_CLIENTS` | `3` | Number of concurrent WebSocket clients |
| `ACTION_INTERVAL` | `10` (web), `15` (mobile), `8` (desktop) | Seconds between actions |

---

## Understanding the Tests

### Client Behavior Patterns

#### Web Application
- **Frequency**: Medium (10 second intervals)
- **Actions**: Dashboard, ticket boards, creation, comments, searches, reports
- **Focus**: General user workflows

#### Android Mobile
- **Frequency**: Slower (15 second intervals)
- **Actions**: Pull to refresh, quick views, mobile comments, notifications
- **Focus**: On-the-go usage patterns

#### Desktop Application
- **Frequency**: Faster (8 second intervals)
- **Actions**: Bulk operations, advanced filters, exports, admin tasks, git integration
- **Focus**: Power user workflows

### WebSocket Event Types

The real-time testing covers these events:
- `ticket.created`
- `ticket.updated`
- `ticket.deleted`
- `ticket.assigned`
- `ticket.commented`
- `project.created`
- `project.updated`
- `sprint.started`
- `sprint.completed`
- `user.mentioned`

---

## Interpreting Results

### Success Criteria

✅ **Complete Success:**
- All 4 phases pass
- All client simulations complete without errors
- WebSocket connections establish successfully
- Real-time events are delivered

⚠️ **Partial Success:**
- Some client simulations fail
- Some API calls return errors
- Check logs for specific failures

❌ **Failure:**
- Organization or project setup fails
- Server is not responding
- Multiple client failures

### Reports

The comprehensive report (`AI_QA_COMPREHENSIVE_REPORT.md`) includes:
1. Executive summary
2. Phase-by-phase results
3. Success rate statistics
4. Output file locations
5. Detailed phase descriptions

---

## Troubleshooting

### Server Not Running

```
Error: Server is not responding
```

**Solution:** Start HelixTrack Core:
```bash
cd Application
./htCore
```

### Organization Already Exists

```
Organization already exists. Skipping setup...
```

**To Reset:**
```bash
rm -rf test-scripts/ai-qa-output/*.json
```

### WebSocket Client Not Found

```
Error: No WebSocket client found!
```

**Solution:** Install websocat or wscat:
```bash
cargo install websocat
# OR
npm install -g wscat
```

### Permission Denied

```
bash: ./ai-qa-comprehensive-test.sh: Permission denied
```

**Solution:** Make scripts executable:
```bash
chmod +x test-scripts/ai-qa-*.sh
```

---

## API Coverage

The test suite covers all 282 API actions across:

### V1 Core Features (144 actions)
- Projects, tickets, users, teams
- Boards, sprints, workflows
- Comments, attachments, labels
- Audit logs, reports

### Phase 1: JIRA Parity (45 actions)
- Priorities, resolutions, versions
- Watchers, filters, custom fields

### Phase 2: Agile Enhancements (62 actions)
- Epics, subtasks, work logs
- Sprint management, burndown charts
- Velocity tracking

### Phase 3: Collaboration (31 actions)
- Notifications, mentions
- Activity streams, dashboards
- Advanced permissions

---

## Real-World Simulation

The test suite simulates 6 months of enterprise development work:

**Month 1-2:** Setup and initial sprints
- Organization creation
- Project setup
- First sprints

**Month 3-4:** Active development
- Multiple parallel sprints
- Heavy ticket creation
- Cross-team collaboration

**Month 5-6:** Delivery and refinement
- Final sprints
- Bug fixes
- Production deployment

---

## Best Practices

1. **Run on Clean Database:** Start with fresh database for accurate testing
2. **Monitor Resources:** Client simulations can be resource-intensive
3. **Review Logs:** Always check logs after failures
4. **Customize Durations:** Adjust durations based on your testing needs
5. **Parallel Execution:** Client simulations run in parallel for efficiency

---

## Advanced Usage

### Long-Running Test

```bash
# 1 hour client simulations, 10 WebSocket clients
CLIENT_DURATION=3600 \
CONCURRENT_CLIENTS=10 \
WS_DURATION=600 \
./ai-qa-comprehensive-test.sh
```

### Stress Testing

```bash
# Multiple iterations
for i in {1..10}; do
    echo "Iteration $i"
    ./ai-qa-comprehensive-test.sh
    sleep 60
done
```

### Custom Client Mix

```bash
# Run only specific clients
SIMULATION_DURATION=600 ./ai-qa-client-webapp.sh &
SIMULATION_DURATION=600 ./ai-qa-client-desktop.sh &
wait
```

---

## Contributing

To extend the AI QA suite:

1. **Add New Client:** Create `ai-qa-client-newtype.sh` based on existing templates
2. **Add New Events:** Extend `ai-qa-websocket-realtime.sh` with new event types
3. **Add New Projects:** Edit `ai-qa-data-projects.json`
4. **Add New Users:** Edit `ai-qa-data-organization.json`

---

## Support

For issues or questions:
- **GitHub Issues**: https://github.com/Helix-Track/Core/issues
- **Documentation**: See `docs/USER_MANUAL.md`
- **Test Reports**: Review generated `AI_QA_COMPREHENSIVE_REPORT.md`

---

**AI QA Test Suite Version:** 1.0.0
**Last Updated:** October 12, 2025
**Compatible with:** HelixTrack Core V3.0+
**Status:** ✅ Production Ready

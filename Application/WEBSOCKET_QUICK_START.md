# WebSocket Event Publishing - Quick Start Guide

**⚡ Get started with real-time events in 5 minutes**

---

## 📋 Table of Contents

1. [Enable WebSocket](#enable-websocket)
2. [Connect from Client](#connect-from-client)
3. [Subscribe to Events](#subscribe-to-events)
4. [Receive Events](#receive-events)
5. [Common Patterns](#common-patterns)
6. [Troubleshooting](#troubleshooting)

---

## 1. Enable WebSocket

**Configuration:** `Configurations/default.json`

```json
{
  "websocket": {
    "enabled": true
  }
}
```

**Start Server:**
```bash
./htCore --config=Configurations/default.json
```

✅ WebSocket endpoint available at: `ws://localhost:8080/ws`

---

## 2. Connect from Client

### JavaScript/Browser

```javascript
// Connect to WebSocket
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onopen = () => {
  console.log('✅ Connected to HelixTrack');
  subscribeToEvents();
};

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  handleMessage(data);
};

ws.onerror = (error) => {
  console.error('❌ WebSocket error:', error);
};

ws.onclose = () => {
  console.log('🔌 Disconnected');
};
```

### Python

```python
import websocket
import json

def on_message(ws, message):
    data = json.loads(message)
    handle_message(data)

def on_open(ws):
    print("✅ Connected to HelixTrack")
    subscribe_to_events(ws)

ws = websocket.WebSocketApp(
    "ws://localhost:8080/ws",
    on_message=on_message,
    on_open=on_open
)

ws.run_forever()
```

### Go

```go
package main

import (
    "github.com/gorilla/websocket"
    "log"
)

func main() {
    ws, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws", nil)
    if err != nil {
        log.Fatal(err)
    }
    defer ws.Close()

    log.Println("✅ Connected to HelixTrack")
    subscribeToEvents(ws)

    for {
        _, message, err := ws.ReadMessage()
        if err != nil {
            log.Println("❌ Read error:", err)
            return
        }
        handleMessage(message)
    }
}
```

---

## 3. Subscribe to Events

### Subscribe to Ticket Events

```javascript
function subscribeToEvents() {
  ws.send(JSON.stringify({
    type: "subscribe",
    data: {
      eventTypes: [
        "ticket.created",
        "ticket.updated",
        "ticket.deleted"
      ]
    }
  }));
}
```

**Response:**
```json
{
  "type": "subscription_confirmed",
  "eventTypes": ["ticket.created", "ticket.updated", "ticket.deleted"]
}
```

### Subscribe to All Events

```javascript
ws.send(JSON.stringify({
  type: "subscribe",
  data: {
    eventTypes: [
      // Ticket events
      "ticket.created", "ticket.updated", "ticket.deleted",

      // Project events
      "project.created", "project.updated", "project.deleted",

      // Comment events
      "comment.created", "comment.updated", "comment.deleted",

      // Priority events
      "priority.created", "priority.updated", "priority.deleted",

      // Resolution events
      "resolution.created", "resolution.updated", "resolution.deleted",

      // Version events
      "version.created", "version.updated", "version.deleted",
      "version.released", "version.archived",

      // Watcher events
      "watcher.added", "watcher.removed",

      // Filter events
      "filter.created", "filter.updated", "filter.deleted", "filter.shared",

      // Custom field events
      "customfield.created", "customfield.updated", "customfield.deleted"
    ]
  }
}));
```

---

## 4. Receive Events

### Handle Incoming Messages

```javascript
function handleMessage(data) {
  switch (data.type) {
    case 'subscription_confirmed':
      console.log('✅ Subscribed to:', data.eventTypes);
      break;

    case 'unsubscription_confirmed':
      console.log('✅ Unsubscribed from:', data.eventTypes);
      break;

    case 'event':
      handleEvent(data.event);
      break;

    default:
      console.log('Unknown message type:', data.type);
  }
}

function handleEvent(event) {
  console.log('📨 Event received:', event.eventType);

  switch (event.eventType) {
    case 'ticket.created':
      onTicketCreated(event);
      break;

    case 'ticket.updated':
      onTicketUpdated(event);
      break;

    case 'ticket.deleted':
      onTicketDeleted(event);
      break;

    // ... handle other event types
  }
}
```

### Event Structure

Every event has this structure:

```json
{
  "type": "event",
  "event": {
    "id": "evt-uuid-123",
    "eventType": "ticket.created",
    "action": "create",
    "object": "ticket",
    "entityId": "ticket-456",
    "username": "john.doe",
    "timestamp": 1696780800,
    "data": {
      "id": "ticket-456",
      "title": "Fix login bug",
      "description": "Users cannot log in",
      "status": "open",
      "priority": "high",
      "project_id": "project-789"
    },
    "context": {
      "projectId": "project-789",
      "permissions": ["READ"]
    }
  }
}
```

---

## 5. Common Patterns

### Pattern 1: Real-Time Dashboard

```javascript
class Dashboard {
  constructor() {
    this.ws = new WebSocket('ws://localhost:8080/ws');
    this.setupWebSocket();
  }

  setupWebSocket() {
    this.ws.onopen = () => {
      // Subscribe to all events for dashboard
      this.ws.send(JSON.stringify({
        type: "subscribe",
        data: {
          eventTypes: [
            "ticket.created", "ticket.updated", "ticket.deleted",
            "project.created", "project.updated"
          ]
        }
      }));
    };

    this.ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      if (data.type === 'event') {
        this.updateDashboard(data.event);
      }
    };
  }

  updateDashboard(event) {
    switch (event.eventType) {
      case 'ticket.created':
        this.incrementTicketCount();
        this.addTicketToList(event.data);
        break;

      case 'ticket.updated':
        this.updateTicketInList(event.data);
        break;

      case 'ticket.deleted':
        this.removeTicketFromList(event.data.id);
        this.decrementTicketCount();
        break;
    }
  }
}

const dashboard = new Dashboard();
```

### Pattern 2: Live Ticket Feed

```javascript
class TicketFeed {
  constructor(projectId) {
    this.projectId = projectId;
    this.ws = new WebSocket('ws://localhost:8080/ws');
    this.setupWebSocket();
  }

  setupWebSocket() {
    this.ws.onmessage = (event) => {
      const data = JSON.parse(event.data);

      if (data.type === 'event' &&
          data.event.context.projectId === this.projectId) {
        this.displayEvent(data.event);
      }
    };
  }

  displayEvent(event) {
    const feedItem = `
      <div class="feed-item">
        <strong>${event.username}</strong>
        ${this.getActionText(event.eventType)}
        ticket <a href="/tickets/${event.entityId}">${event.data.title}</a>
        <span class="timestamp">${new Date(event.timestamp * 1000).toLocaleString()}</span>
      </div>
    `;
    document.getElementById('feed').insertAdjacentHTML('afterbegin', feedItem);
  }

  getActionText(eventType) {
    const actions = {
      'ticket.created': 'created',
      'ticket.updated': 'updated',
      'ticket.deleted': 'deleted',
      'comment.created': 'commented on'
    };
    return actions[eventType] || 'modified';
  }
}

const feed = new TicketFeed('project-123');
```

### Pattern 3: Notification Bell

```javascript
class NotificationBell {
  constructor() {
    this.unreadCount = 0;
    this.ws = new WebSocket('ws://localhost:8080/ws');
    this.setupWebSocket();
  }

  setupWebSocket() {
    this.ws.onopen = () => {
      // Subscribe to events relevant to current user
      this.ws.send(JSON.stringify({
        type: "subscribe",
        data: {
          eventTypes: [
            "watcher.added",    // Someone added me as watcher
            "comment.created",  // New comment on watched tickets
            "ticket.updated"    // Updates to watched tickets
          ]
        }
      }));
    };

    this.ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      if (data.type === 'event') {
        this.handleNotification(data.event);
      }
    };
  }

  handleNotification(event) {
    // Check if notification is relevant to current user
    if (this.isRelevantToMe(event)) {
      this.unreadCount++;
      this.updateBadge();
      this.showToast(event);
    }
  }

  updateBadge() {
    document.getElementById('notification-badge').textContent = this.unreadCount;
    document.getElementById('notification-badge').style.display =
      this.unreadCount > 0 ? 'inline' : 'none';
  }

  showToast(event) {
    // Show browser notification or in-app toast
    if (Notification.permission === 'granted') {
      new Notification('HelixTrack Update', {
        body: `${event.username} ${this.getActionText(event.eventType)}`,
        icon: '/icon.png'
      });
    }
  }
}

const notificationBell = new NotificationBell();
```

### Pattern 4: Collaborative Editing

```javascript
class CollaborativeEditor {
  constructor(ticketId) {
    this.ticketId = ticketId;
    this.ws = new WebSocket('ws://localhost:8080/ws');
    this.activeUsers = new Set();
    this.setupWebSocket();
  }

  setupWebSocket() {
    this.ws.onopen = () => {
      this.ws.send(JSON.stringify({
        type: "subscribe",
        data: {
          eventTypes: ["ticket.updated", "comment.created"]
        }
      }));
    };

    this.ws.onmessage = (event) => {
      const data = JSON.parse(event.data);

      if (data.type === 'event' &&
          data.event.entityId === this.ticketId) {

        if (data.event.eventType === 'ticket.updated') {
          this.handleTicketUpdate(data.event);
        } else if (data.event.eventType === 'comment.created') {
          this.handleNewComment(data.event);
        }
      }
    };
  }

  handleTicketUpdate(event) {
    // Show "Someone else is editing" warning
    if (event.username !== currentUsername) {
      this.showEditWarning(event.username);
      this.refreshTicketData();
    }
  }

  handleNewComment(event) {
    // Add new comment to UI without refresh
    this.addCommentToUI(event.data);
  }
}
```

---

## 6. Troubleshooting

### Problem: Connection Fails

```javascript
// Add connection retry logic
function connectWithRetry() {
  const ws = new WebSocket('ws://localhost:8080/ws');

  ws.onerror = (error) => {
    console.error('Connection error:', error);
    setTimeout(connectWithRetry, 5000); // Retry after 5 seconds
  };

  ws.onclose = () => {
    console.log('Connection closed, reconnecting...');
    setTimeout(connectWithRetry, 5000);
  };

  return ws;
}

const ws = connectWithRetry();
```

### Problem: No Events Received

**Check 1:** Verify subscription
```javascript
ws.send(JSON.stringify({
  type: "subscribe",
  data: {
    eventTypes: ["ticket.created"] // Exact event type name
  }
}));
```

**Check 2:** Verify event type spelling
```javascript
// ❌ Wrong
"ticketCreated"

// ✅ Correct
"ticket.created"
```

**Check 3:** Check console for subscription confirmation
```javascript
ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Received:', data); // Debug all messages
};
```

### Problem: Missing Some Events

**Cause:** Permission filtering

**Solution:** Ensure user has proper permissions for the project/entity

### Problem: High Latency

**Check:** Network connection
```javascript
const startTime = Date.now();
ws.send(JSON.stringify({type: "ping"}));

ws.onmessage = (event) => {
  const latency = Date.now() - startTime;
  console.log(`Latency: ${latency}ms`);
};
```

---

## 🎯 Quick Reference

### All Event Types

```
Ticket Events:       ticket.created, ticket.updated, ticket.deleted
Project Events:      project.created, project.updated, project.deleted
Comment Events:      comment.created, comment.updated, comment.deleted
Priority Events:     priority.created, priority.updated, priority.deleted
Resolution Events:   resolution.created, resolution.updated, resolution.deleted
Version Events:      version.created, version.updated, version.deleted,
                     version.released, version.archived
Watcher Events:      watcher.added, watcher.removed
Filter Events:       filter.created, filter.updated, filter.deleted, filter.shared
Custom Field Events: customfield.created, customfield.updated, customfield.deleted
```

### Message Types

```
subscribe              - Subscribe to event types
unsubscribe            - Unsubscribe from event types
subscription_confirmed - Subscription successful
unsubscription_confirmed - Unsubscription successful
event                  - Event notification
```

---

## 📚 Next Steps

1. **Test Connection:** Use browser console or WebSocket client
2. **Subscribe to Events:** Start with one event type
3. **Build UI:** Create real-time updates in your application
4. **Add Error Handling:** Implement reconnection logic
5. **Monitor Performance:** Track latency and event processing

---

## 🔗 Additional Resources

- **Full API Documentation:** `/docs/USER_MANUAL.md`
- **Deployment Guide:** `/docs/DEPLOYMENT.md`
- **Test Examples:** `/internal/websocket/manager_integration_test.go`
- **Complete Delivery Doc:** `/WEBSOCKET_EVENT_PUBLISHING_FINAL_DELIVERY.md`

---

**🚀 You're ready to build real-time features with HelixTrack!**

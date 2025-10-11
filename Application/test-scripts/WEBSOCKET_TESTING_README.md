# WebSocket Testing Guide

This directory contains tools for testing the HelixTrack Core WebSocket event notification system.

## Available Test Tools

### 1. Bash Test Script (`test-websocket.sh`)

Automated test script that:
- Registers a test user
- Authenticates and gets JWT token
- Tests WebSocket stats endpoint
- Attempts WebSocket connection (if websocat/wscat installed)

**Usage:**

```bash
# Default (localhost:8080)
./test-websocket.sh

# Custom server
BASE_URL=http://your-server:8080 WS_URL=ws://your-server:8080/ws ./test-websocket.sh

# Custom credentials
USERNAME=myuser PASSWORD=mypass ./test-websocket.sh
```

**Prerequisites:**
- curl (required)
- websocat or wscat (optional, for interactive WebSocket testing)

**Install WebSocket clients:**
```bash
# websocat (recommended)
cargo install websocat

# OR wscat
npm install -g wscat
```

### 2. HTML/JavaScript Client (`websocket-client.html`)

Interactive web-based WebSocket client with GUI.

**Features:**
- User authentication interface
- WebSocket connection management
- Event subscription configuration
- Real-time event display
- Event filtering and logging

**Usage:**

1. Open `websocket-client.html` in your web browser
2. Enter your server URL and credentials
3. Click "Authenticate" to get JWT token
4. Click "Connect" to establish WebSocket connection
5. Configure your event subscriptions
6. Click "Subscribe" to start receiving events

**Or serve it locally:**
```bash
python3 -m http.server 8000
# Then open http://localhost:8000/websocket-client.html
```

## Testing Workflow

### Basic Test Flow

1. **Start the server with WebSocket enabled:**
   ```bash
   cd Application
   ./htCore --config=Configurations/dev_with_websocket.json
   ```

2. **Run the bash test script:**
   ```bash
   cd test-scripts
   ./test-websocket.sh
   ```

3. **Open the HTML client in your browser:**
   ```bash
   # Direct file open
   open websocket-client.html  # macOS
   xdg-open websocket-client.html  # Linux
   start websocket-client.html  # Windows

   # Or serve it
   python3 -m http.server 8000
   ```

### Testing Event Broadcasting

**Terminal 1 - Start Server:**
```bash
cd Application
./htCore --config=Configurations/dev_with_websocket.json
```

**Terminal 2 - Connect WebSocket Client:**
```bash
cd test-scripts

# Get JWT token
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"testpass"}' \
  | grep -o '"token":"[^"]*' | cut -d'"' -f4)

# Connect and subscribe
echo '{"type":"subscribe","data":{"eventTypes":["ticket.created","ticket.updated"]}}' \
  | websocat "ws://localhost:8080/ws?token=$TOKEN"
```

**Terminal 3 - Trigger Events:**
```bash
# Create a ticket (this will trigger an event)
curl -X POST http://localhost:8080/do \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"action":"create","object":"ticket","data":{"title":"Test Ticket","projectId":"project-123"}}'
```

You should see the event appear in Terminal 2!

## Event Types

You can subscribe to the following event types:

**Ticket Events:**
- `ticket.created`
- `ticket.updated`
- `ticket.deleted`
- `ticket.read` (optional)

**Project Events:**
- `project.created`
- `project.updated`
- `project.deleted`
- `project.read` (optional)

**Comment Events:**
- `comment.created`
- `comment.updated`
- `comment.deleted`

**Priority/Resolution/Version Events:**
- `priority.created`, `priority.updated`, `priority.deleted`
- `resolution.created`, `resolution.updated`, `resolution.deleted`
- `version.created`, `version.updated`, `version.deleted`
- `version.released`, `version.archived`

**System Events:**
- `connection.established`
- `connection.closed`
- `system.error`

## Subscription Examples

### Subscribe to All Ticket Events

```json
{
  "type": "subscribe",
  "data": {
    "eventTypes": [
      "ticket.created",
      "ticket.updated",
      "ticket.deleted"
    ]
  }
}
```

### Subscribe to Project-Specific Events

```json
{
  "type": "subscribe",
  "data": {
    "entityTypes": ["ticket", "project"],
    "filters": {
      "projectId": "project-123"
    }
  }
}
```

### Subscribe to Specific Entity IDs

```json
{
  "type": "subscribe",
  "data": {
    "entityIds": ["ticket-123", "ticket-456"]
  }
}
```

### Subscribe with Read Events

```json
{
  "type": "subscribe",
  "data": {
    "eventTypes": ["ticket.created", "ticket.updated"],
    "includeReads": true
  }
}
```

## Message Protocol

### Client → Server Messages

**Subscribe:**
```json
{
  "type": "subscribe",
  "data": {
    "eventTypes": ["ticket.created"],
    "entityTypes": ["ticket"],
    "entityIds": ["ticket-123"],
    "filters": {
      "projectId": "project-456"
    },
    "includeReads": false
  }
}
```

**Unsubscribe:**
```json
{
  "type": "unsubscribe"
}
```

**Ping:**
```json
{
  "type": "ping"
}
```

### Server → Client Messages

**Event:**
```json
{
  "type": "event",
  "action": "create",
  "eventId": "event-123",
  "data": {
    "event": {
      "id": "event-123",
      "type": "ticket.created",
      "action": "create",
      "object": "ticket",
      "entityId": "ticket-456",
      "username": "john.doe",
      "timestamp": "2025-10-11T10:30:00Z",
      "data": {
        "title": "New Ticket"
      }
    }
  }
}
```

**Acknowledgment:**
```json
{
  "type": "ack",
  "action": "subscribe",
  "data": {
    "success": true,
    "message": "Subscription updated"
  }
}
```

**Error:**
```json
{
  "type": "error",
  "error": "Invalid message format"
}
```

**Pong:**
```json
{
  "type": "pong",
  "data": {
    "time": "2025-10-11T10:30:00Z"
  }
}
```

## Troubleshooting

### Connection Refused

**Problem:** Cannot connect to WebSocket

**Solutions:**
1. Check if server is running: `curl http://localhost:8080/health`
2. Verify WebSocket is enabled in configuration
3. Check firewall settings
4. Verify port is not already in use

### Authentication Failed

**Problem:** Invalid or expired JWT token

**Solutions:**
1. Re-authenticate to get a new token
2. Check token expiration time
3. Verify authentication service is enabled
4. Check username/password are correct

### No Events Received

**Problem:** Connected but not receiving events

**Solutions:**
1. Verify subscription was sent correctly
2. Check event filters match the events being triggered
3. Verify permission settings allow access to events
4. Check server logs for errors
5. Ensure events are actually being triggered (create a ticket)

### WebSocket Disconnects Frequently

**Problem:** Connection drops after a short time

**Solutions:**
1. Check network stability
2. Verify reverse proxy timeout settings (if applicable)
3. Check client ping/pong implementation
4. Review server logs for errors
5. Check `pongWait` configuration setting

## Performance Testing

### Load Test with Multiple Clients

```bash
# Terminal 1-10: Connect multiple clients
for i in {1..10}; do
  (echo '{"type":"subscribe","data":{"eventTypes":["ticket.created"]}}' \
    | websocat "ws://localhost:8080/ws?token=$TOKEN" &)
done

# Terminal 11: Trigger many events
for i in {1..100}; do
  curl -s -X POST http://localhost:8080/do \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"action\":\"create\",\"object\":\"ticket\",\"data\":{\"title\":\"Ticket $i\"}}" > /dev/null
done
```

### Check WebSocket Stats

```bash
curl http://localhost:8080/ws/stats | jq
```

## Security Testing

### Test JWT Authentication

```bash
# Should fail without token
websocat ws://localhost:8080/ws

# Should fail with invalid token
websocat "ws://localhost:8080/ws?token=invalid"

# Should succeed with valid token
websocat "ws://localhost:8080/ws?token=$VALID_TOKEN"
```

### Test Permission Filtering

Create events with different permission contexts and verify clients only receive events they have permissions for.

## Integration with Other Tools

### Postman

1. Create a new WebSocket request
2. Set URL: `ws://localhost:8080/ws?token={{jwtToken}}`
3. Connect
4. Send subscription message
5. Trigger events via REST API
6. Observe events in WebSocket messages

### Python Client

```python
import websockets
import asyncio
import json

async def test_websocket():
    uri = "ws://localhost:8080/ws?token=YOUR_JWT_TOKEN"
    async with websockets.connect(uri) as websocket:
        # Subscribe
        await websocket.send(json.dumps({
            "type": "subscribe",
            "data": {
                "eventTypes": ["ticket.created"]
            }
        }))

        # Receive events
        async for message in websocket:
            print(f"Received: {message}")

asyncio.run(test_websocket())
```

## Next Steps

- Review `EVENT_INTEGRATION_PATTERN.md` for handler integration
- Check `WEBSOCKET_IMPLEMENTATION_SUMMARY.md` for architecture details
- See `USER_MANUAL.md` for complete API documentation
- Review `DEPLOYMENT.md` for production deployment

## Support

For issues or questions:
1. Check server logs: `/tmp/htCoreLogs/htCore.log`
2. Review configuration: `Configurations/dev_with_websocket.json`
3. Test with HTML client first to isolate issues
4. Verify network connectivity and firewall settings

# HelixTrack Chat Service API Documentation

Complete API reference for the HelixTrack Chat microservice.

## Table of Contents

- [Overview](#overview)
- [Authentication](#authentication)
- [Request/Response Format](#requestresponse-format)
- [Error Handling](#error-handling)
- [Rate Limiting](#rate-limiting)
- [Chat Room API](#chat-room-api)
- [Message API](#message-api)
- [Participant API](#participant-api)
- [Real-Time API](#real-time-api)
- [System Endpoints](#system-endpoints)
- [WebSocket Events](#websocket-events)

## Overview

The Chat Service provides a unified REST API through the `/do` endpoint with action-based routing. All actions use HTTP POST and return JSON responses.

**Base URL**: `http://localhost:9090/api` (development)
**Production URL**: `https://chat.helixtrack.io/api`

**Available Actions**: 31 total
- Chat Room: 6 actions
- Messages: 10 actions
- Participants: 6 actions
- Real-Time: 9 actions

## Authentication

All API endpoints (except `/health` and `/version`) require JWT authentication.

### JWT Token Format

```json
{
  "sub": "authentication",
  "name": "John Doe",
  "username": "johndoe",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "role": "admin",
  "permissions": "READ|CREATE|UPDATE|DELETE",
  "htCoreAddress": "http://core-service:8080",
  "exp": 1234567890,
  "iat": 1234567890
}
```

### Authentication Methods

**1. Authorization Header (Recommended)**
```bash
Authorization: Bearer YOUR_JWT_TOKEN
```

**2. Query Parameter**
```bash
?jwt=YOUR_JWT_TOKEN
```

**3. Request Body**
```json
{
  "action": "chatRoomList",
  "jwt": "YOUR_JWT_TOKEN"
}
```

### Authentication Errors

```json
{
  "errorCode": 1003,
  "errorMessage": "Invalid or expired JWT token",
  "data": null
}
```

## Request/Response Format

### Request Structure

```json
{
  "action": "string",      // Required: Action name
  "jwt": "string",         // Optional if in header/query
  "data": {                // Action-specific data
    "key": "value"
  }
}
```

### Response Structure

```json
{
  "errorCode": -1,                  // -1 = success, positive = error
  "errorMessage": "string",         // Error description (empty on success)
  "data": {}                        // Response data (null on error)
}
```

### Success Response Example

```json
{
  "errorCode": -1,
  "errorMessage": "",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Project Alpha Chat",
    "created_at": "2025-10-17T10:30:00Z"
  }
}
```

### Error Response Example

```json
{
  "errorCode": 3000,
  "errorMessage": "Chat room not found",
  "data": null
}
```

## Error Handling

### Error Code Ranges

- **-1**: Success (no error)
- **1000-1999**: Request errors
- **2000-2999**: System errors
- **3000-3999**: Entity errors
- **4000-4999**: Rate limiting

### Common Error Codes

| Code | Message | Description |
|------|---------|-------------|
| -1 | Success | Operation completed successfully |
| 1000 | Invalid request format | Malformed JSON or missing fields |
| 1001 | Missing required parameter | Required field not provided |
| 1002 | Invalid parameter value | Parameter value is invalid |
| 1003 | Invalid JWT token | Token is invalid or expired |
| 2000 | Database error | Database operation failed |
| 2001 | Internal server error | Unexpected server error |
| 3000 | Entity not found | Requested entity doesn't exist |
| 3001 | Entity already exists | Entity with ID already exists |
| 3002 | Forbidden | Insufficient permissions |
| 4000 | Rate limit exceeded | Too many requests |

## Rate Limiting

**Default Limits:**
- 10 requests per second per IP address
- Burst capacity: 20 requests
- Cleanup interval: 5 minutes

**Rate Limit Headers:**
```
X-RateLimit-Limit: 10
X-RateLimit-Remaining: 7
X-RateLimit-Reset: 1634567890
```

**Rate Limit Response:**
```json
{
  "errorCode": 4000,
  "errorMessage": "Too many requests. Please try again later.",
  "data": null
}
```

## Chat Room API

### 1. chatRoomCreate

Create a new chat room.

**Action**: `chatRoomCreate`

**Permissions**: Authenticated users

**Request:**
```json
{
  "action": "chatRoomCreate",
  "data": {
    "name": "string",              // Required: Room name (1-255 chars)
    "description": "string",       // Optional: Room description
    "type": "string",              // Required: direct|group|channel|private
    "is_private": boolean,         // Optional: Default false
    "entity_type": "string",       // Optional: user|team|project|ticket|etc
    "entity_id": "uuid",           // Optional: Associated entity ID
    "metadata": {}                 // Optional: Custom metadata (JSONB)
  }
}
```

**Response:**
```json
{
  "errorCode": -1,
  "errorMessage": "",
  "data": {
    "id": "uuid",
    "name": "Project Alpha Chat",
    "description": "Discussion for Project Alpha",
    "type": "group",
    "is_private": false,
    "entity_type": "project",
    "entity_id": "uuid",
    "created_by": "uuid",
    "created_at": "2025-10-17T10:30:00Z",
    "updated_at": "2025-10-17T10:30:00Z",
    "deleted": false,
    "metadata": {}
  }
}
```

**Notes:**
- Creator is automatically added as room owner
- Room types: `direct` (1-on-1), `group` (multiple users), `channel` (public), `private` (invite-only)

---

### 2. chatRoomRead

Get details of a specific chat room.

**Action**: `chatRoomRead`

**Permissions**: Room participants only

**Request:**
```json
{
  "action": "chatRoomRead",
  "data": {
    "id": "uuid"                   // Required: Room ID
  }
}
```

**Response:**
```json
{
  "errorCode": -1,
  "errorMessage": "",
  "data": {
    "id": "uuid",
    "name": "Project Alpha Chat",
    "description": "Discussion for Project Alpha",
    "type": "group",
    "is_private": false,
    "entity_type": "project",
    "entity_id": "uuid",
    "created_by": "uuid",
    "created_at": "2025-10-17T10:30:00Z",
    "updated_at": "2025-10-17T10:30:00Z",
    "deleted": false,
    "metadata": {}
  }
}
```

---

### 3. chatRoomList

List all chat rooms for the authenticated user.

**Action**: `chatRoomList`

**Permissions**: Authenticated users

**Request:**
```json
{
  "action": "chatRoomList",
  "data": {
    "limit": 20,                   // Optional: Max items (default 20, max 100)
    "offset": 0,                   // Optional: Pagination offset (default 0)
    "type": "string",              // Optional: Filter by type
    "entity_type": "string",       // Optional: Filter by entity type
    "entity_id": "uuid"            // Optional: Filter by entity ID
  }
}
```

**Response:**
```json
{
  "errorCode": -1,
  "errorMessage": "",
  "data": {
    "items": [
      {
        "id": "uuid",
        "name": "Project Alpha Chat",
        "type": "group",
        "created_at": "2025-10-17T10:30:00Z"
      }
    ],
    "pagination": {
      "total": 42,
      "limit": 20,
      "offset": 0,
      "hasMore": true
    }
  }
}
```

---

### 4. chatRoomUpdate

Update chat room details.

**Action**: `chatRoomUpdate`

**Permissions**: Room owner or admin

**Request:**
```json
{
  "action": "chatRoomUpdate",
  "data": {
    "id": "uuid",                  // Required: Room ID
    "name": "string",              // Optional: New name
    "description": "string",       // Optional: New description
    "is_private": boolean,         // Optional: Change privacy
    "metadata": {}                 // Optional: Update metadata
  }
}
```

**Response:**
```json
{
  "errorCode": -1,
  "errorMessage": "",
  "data": {
    "id": "uuid",
    "name": "Updated Room Name",
    "description": "Updated description",
    "updated_at": "2025-10-17T11:00:00Z"
  }
}
```

---

### 5. chatRoomDelete

Delete a chat room (soft delete).

**Action**: `chatRoomDelete`

**Permissions**: Room owner only

**Request:**
```json
{
  "action": "chatRoomDelete",
  "data": {
    "id": "uuid"                   // Required: Room ID
  }
}
```

**Response:**
```json
{
  "errorCode": -1,
  "errorMessage": "",
  "data": {
    "id": "uuid",
    "deleted": true,
    "deleted_at": "2025-10-17T11:30:00Z"
  }
}
```

---

### 6. chatRoomGetByEntity

Get chat room for a specific entity.

**Action**: `chatRoomGetByEntity`

**Permissions**: Authenticated users

**Request:**
```json
{
  "action": "chatRoomGetByEntity",
  "data": {
    "entity_type": "string",       // Required: Entity type
    "entity_id": "uuid"            // Required: Entity ID
  }
}
```

**Response:**
```json
{
  "errorCode": -1,
  "errorMessage": "",
  "data": {
    "id": "uuid",
    "name": "Ticket #123 Discussion",
    "entity_type": "ticket",
    "entity_id": "uuid",
    "created_at": "2025-10-17T10:30:00Z"
  }
}
```

## Message API

### 1. messageSend

Send a new message to a chat room.

**Action**: `messageSend`

**Permissions**: Room participants

**Request:**
```json
{
  "action": "messageSend",
  "data": {
    "chat_room_id": "uuid",        // Required: Room ID
    "content": "string",           // Required: Message content (1-10000 chars)
    "type": "string",              // Optional: text|image|file|system (default: text)
    "content_format": "string",    // Optional: plain|markdown (default: plain)
    "metadata": {}                 // Optional: Custom metadata
  }
}
```

**Response:**
```json
{
  "errorCode": -1,
  "errorMessage": "",
  "data": {
    "id": "uuid",
    "chat_room_id": "uuid",
    "sender_id": "uuid",
    "type": "text",
    "content": "Hello, world!",
    "content_format": "plain",
    "is_edited": false,
    "is_pinned": false,
    "created_at": "2025-10-17T10:30:00Z",
    "metadata": {}
  }
}
```

---

### 2. messageReply

Reply to an existing message (creates thread).

**Action**: `messageReply`

**Permissions**: Room participants

**Request:**
```json
{
  "action": "messageReply",
  "data": {
    "chat_room_id": "uuid",        // Required: Room ID
    "parent_id": "uuid",           // Required: Parent message ID
    "content": "string",           // Required: Reply content
    "type": "string",              // Optional: Message type
    "content_format": "string"     // Optional: Content format
  }
}
```

**Response:**
```json
{
  "errorCode": -1,
  "errorMessage": "",
  "data": {
    "id": "uuid",
    "chat_room_id": "uuid",
    "sender_id": "uuid",
    "parent_id": "uuid",           // Thread parent
    "content": "This is a reply",
    "created_at": "2025-10-17T10:32:00Z"
  }
}
```

---

### 3. messageQuote

Quote a message in a new message.

**Action**: `messageQuote`

**Permissions**: Room participants

**Request:**
```json
{
  "action": "messageQuote",
  "data": {
    "chat_room_id": "uuid",        // Required: Room ID
    "quoted_message_id": "uuid",   // Required: Message to quote
    "content": "string"            // Required: Your message
  }
}
```

**Response:**
```json
{
  "errorCode": -1,
  "errorMessage": "",
  "data": {
    "id": "uuid",
    "chat_room_id": "uuid",
    "sender_id": "uuid",
    "quoted_message_id": "uuid",   // Quoted message
    "content": "I agree with this",
    "created_at": "2025-10-17T10:33:00Z"
  }
}
```

---

### 4. messageList

List messages in a chat room (paginated).

**Action**: `messageList`

**Permissions**: Room participants

**Request:**
```json
{
  "action": "messageList",
  "data": {
    "chat_room_id": "uuid",        // Required: Room ID
    "limit": 50,                   // Optional: Max items (default 50, max 100)
    "offset": 0,                   // Optional: Pagination offset
    "order": "string"              // Optional: asc|desc (default: desc)
  }
}
```

**Response:**
```json
{
  "errorCode": -1,
  "errorMessage": "",
  "data": {
    "items": [
      {
        "id": "uuid",
        "sender_id": "uuid",
        "content": "Hello",
        "created_at": "2025-10-17T10:30:00Z"
      }
    ],
    "pagination": {
      "total": 150,
      "limit": 50,
      "offset": 0,
      "hasMore": true
    }
  }
}
```

---

### 5. messageSearch

Full-text search for messages.

**Action**: `messageSearch`

**Permissions**: Room participants

**Request:**
```json
{
  "action": "messageSearch",
  "data": {
    "chat_room_id": "uuid",        // Required: Room ID
    "query": "string",             // Required: Search query
    "limit": 20,                   // Optional: Max results
    "offset": 0                    // Optional: Pagination
  }
}
```

**Response:**
```json
{
  "errorCode": -1,
  "errorMessage": "",
  "data": {
    "items": [
      {
        "id": "uuid",
        "content": "This matches the search query",
        "created_at": "2025-10-17T09:15:00Z",
        "relevance": 0.95          // Search relevance score
      }
    ],
    "pagination": {
      "total": 5,
      "limit": 20,
      "offset": 0,
      "hasMore": false
    }
  }
}
```

---

### 6. messageRead

Get a single message by ID.

**Action**: `messageRead`

**Request:**
```json
{
  "action": "messageRead",
  "data": {
    "id": "uuid"                   // Required: Message ID
  }
}
```

---

### 7. messageUpdate

Edit your own message.

**Action**: `messageUpdate`

**Permissions**: Message author only

**Request:**
```json
{
  "action": "messageUpdate",
  "data": {
    "id": "uuid",                  // Required: Message ID
    "content": "string"            // Required: New content
  }
}
```

**Response:**
```json
{
  "errorCode": -1,
  "errorMessage": "",
  "data": {
    "id": "uuid",
    "content": "Updated content",
    "is_edited": true,
    "edited_at": "2025-10-17T11:00:00Z"
  }
}
```

---

### 8. messageDelete

Delete a message (soft delete).

**Action**: `messageDelete`

**Permissions**: Message author or room admin

**Request:**
```json
{
  "action": "messageDelete",
  "data": {
    "id": "uuid"                   // Required: Message ID
  }
}
```

---

### 9. messagePin

Pin a message (appears at top of room).

**Action**: `messagePin`

**Permissions**: Room admin or moderator

**Request:**
```json
{
  "action": "messagePin",
  "data": {
    "id": "uuid"                   // Required: Message ID
  }
}
```

**Response:**
```json
{
  "errorCode": -1,
  "errorMessage": "",
  "data": {
    "id": "uuid",
    "is_pinned": true,
    "pinned_at": "2025-10-17T11:00:00Z",
    "pinned_by": "uuid"
  }
}
```

---

### 10. messageUnpin

Unpin a message.

**Action**: `messageUnpin`

**Permissions**: Room admin or moderator

**Request:**
```json
{
  "action": "messageUnpin",
  "data": {
    "id": "uuid"                   // Required: Message ID
  }
}
```

## Participant API

### 1. participantAdd

Add a user to a chat room.

**Action**: `participantAdd`

**Permissions**: Room owner, admin, or moderator

**Request:**
```json
{
  "action": "participantAdd",
  "data": {
    "chat_room_id": "uuid",        // Required: Room ID
    "user_id": "uuid",             // Required: User to add
    "role": "string"               // Optional: owner|admin|moderator|member|guest (default: member)
  }
}
```

**Response:**
```json
{
  "errorCode": -1,
  "errorMessage": "",
  "data": {
    "chat_room_id": "uuid",
    "user_id": "uuid",
    "role": "member",
    "joined_at": "2025-10-17T11:00:00Z"
  }
}
```

---

### 2. participantRemove

Remove a user from a chat room.

**Action**: `participantRemove`

**Permissions**: Room owner/admin (or self-removal)

**Request:**
```json
{
  "action": "participantRemove",
  "data": {
    "chat_room_id": "uuid",        // Required: Room ID
    "user_id": "uuid"              // Required: User to remove
  }
}
```

---

### 3. participantList

List all participants in a room.

**Action**: `participantList`

**Request:**
```json
{
  "action": "participantList",
  "data": {
    "chat_room_id": "uuid"         // Required: Room ID
  }
}
```

**Response:**
```json
{
  "errorCode": -1,
  "errorMessage": "",
  "data": {
    "items": [
      {
        "user_id": "uuid",
        "role": "owner",
        "joined_at": "2025-10-17T10:30:00Z",
        "is_muted": false
      },
      {
        "user_id": "uuid",
        "role": "member",
        "joined_at": "2025-10-17T10:35:00Z",
        "is_muted": false
      }
    ]
  }
}
```

---

### 4. participantUpdateRole

Change a participant's role.

**Action**: `participantUpdateRole`

**Permissions**: Room owner or admin

**Request:**
```json
{
  "action": "participantUpdateRole",
  "data": {
    "chat_room_id": "uuid",        // Required: Room ID
    "user_id": "uuid",             // Required: User ID
    "role": "string"               // Required: New role
  }
}
```

---

### 5. participantMute

Mute a participant (cannot send messages).

**Action**: `participantMute`

**Permissions**: Room moderator or admin

**Request:**
```json
{
  "action": "participantMute",
  "data": {
    "chat_room_id": "uuid",        // Required: Room ID
    "user_id": "uuid"              // Required: User to mute
  }
}
```

---

### 6. participantUnmute

Unmute a participant.

**Action**: `participantUnmute`

**Permissions**: Room moderator or admin

**Request:**
```json
{
  "action": "participantUnmute",
  "data": {
    "chat_room_id": "uuid",        // Required: Room ID
    "user_id": "uuid"              // Required: User to unmute
  }
}
```

## Real-Time API

### Typing Indicators

#### typingStart

Indicate that user is typing.

**Action**: `typingStart`

**Request:**
```json
{
  "action": "typingStart",
  "data": {
    "chat_room_id": "uuid"         // Required: Room ID
  }
}
```

**Notes:**
- Automatically expires after 5 seconds
- Triggers WebSocket event to other participants

---

#### typingStop

Stop typing indicator.

**Action**: `typingStop`

**Request:**
```json
{
  "action": "typingStop",
  "data": {
    "chat_room_id": "uuid"         // Required: Room ID
  }
}
```

---

### Presence

#### presenceUpdate

Update user online status.

**Action**: `presenceUpdate`

**Request:**
```json
{
  "action": "presenceUpdate",
  "data": {
    "status": "string"             // Required: online|offline|away|busy|dnd
  }
}
```

**Response:**
```json
{
  "errorCode": -1,
  "errorMessage": "",
  "data": {
    "user_id": "uuid",
    "status": "online",
    "last_seen": "2025-10-17T11:00:00Z"
  }
}
```

---

#### presenceGet

Get user's presence status.

**Action**: `presenceGet`

**Request:**
```json
{
  "action": "presenceGet",
  "data": {
    "user_id": "uuid"              // Required: User ID
  }
}
```

---

### Read Receipts

#### readReceiptMark

Mark message as read.

**Action**: `readReceiptMark`

**Request:**
```json
{
  "action": "readReceiptMark",
  "data": {
    "message_id": "uuid"           // Required: Message ID
  }
}
```

---

#### readReceiptGet

Get read receipts for a message.

**Action**: `readReceiptGet`

**Request:**
```json
{
  "action": "readReceiptGet",
  "data": {
    "message_id": "uuid"           // Required: Message ID
  }
}
```

**Response:**
```json
{
  "errorCode": -1,
  "errorMessage": "",
  "data": {
    "items": [
      {
        "user_id": "uuid",
        "read_at": "2025-10-17T10:35:00Z"
      }
    ]
  }
}
```

---

### Reactions

#### reactionAdd

Add emoji reaction to a message.

**Action**: `reactionAdd`

**Request:**
```json
{
  "action": "reactionAdd",
  "data": {
    "message_id": "uuid",          // Required: Message ID
    "emoji": "string"              // Required: Emoji (e.g., "👍", "❤️")
  }
}
```

---

#### reactionRemove

Remove your reaction from a message.

**Action**: `reactionRemove`

**Request:**
```json
{
  "action": "reactionRemove",
  "data": {
    "message_id": "uuid",          // Required: Message ID
    "emoji": "string"              // Required: Emoji to remove
  }
}
```

---

#### reactionList

List all reactions on a message.

**Action**: `reactionList`

**Request:**
```json
{
  "action": "reactionList",
  "data": {
    "message_id": "uuid"           // Required: Message ID
  }
}
```

**Response:**
```json
{
  "errorCode": -1,
  "errorMessage": "",
  "data": {
    "items": [
      {
        "emoji": "👍",
        "count": 5,
        "users": ["uuid1", "uuid2", "uuid3"]
      },
      {
        "emoji": "❤️",
        "count": 2,
        "users": ["uuid4", "uuid5"]
      }
    ]
  }
}
```

---

### Attachments

#### attachmentUpload

Attach a file to a message.

**Action**: `attachmentUpload`

**Request:**
```json
{
  "action": "attachmentUpload",
  "data": {
    "message_id": "uuid",          // Required: Message ID
    "file_name": "string",         // Required: File name
    "file_size": integer,          // Required: File size in bytes
    "mime_type": "string",         // Required: MIME type
    "storage_url": "string",       // Required: URL to file
    "metadata": {}                 // Optional: File metadata
  }
}
```

**Response:**
```json
{
  "errorCode": -1,
  "errorMessage": "",
  "data": {
    "id": "uuid",
    "message_id": "uuid",
    "file_name": "document.pdf",
    "file_size": 1024000,
    "mime_type": "application/pdf",
    "storage_url": "https://storage.example.com/files/document.pdf",
    "uploaded_at": "2025-10-17T11:00:00Z"
  }
}
```

---

#### attachmentDelete

Delete an attachment.

**Action**: `attachmentDelete`

**Request:**
```json
{
  "action": "attachmentDelete",
  "data": {
    "id": "uuid"                   // Required: Attachment ID
  }
}
```

---

#### attachmentList

List attachments for a message.

**Action**: `attachmentList`

**Request:**
```json
{
  "action": "attachmentList",
  "data": {
    "message_id": "uuid"           // Required: Message ID
  }
}
```

## System Endpoints

### Health Check

Check service health status.

**Endpoint**: `GET /health`

**Authentication**: None required

**Response:**
```json
{
  "status": "healthy",
  "database": "connected",
  "timestamp": "2025-10-17T11:00:00Z"
}
```

---

### Version

Get service version information.

**Endpoint**: `GET /version`

**Authentication**: None required

**Response:**
```json
{
  "version": "1.0.0",
  "buildTime": "2025-10-17T09:00:00Z",
  "gitCommit": "abc123def456",
  "goVersion": "go1.22.1"
}
```

## WebSocket Events

### Connection

```
wss://chat.helixtrack.io/ws?jwt=YOUR_JWT_TOKEN
```

### Event Types

#### message_new

Triggered when a new message is sent.

```json
{
  "type": "message_new",
  "data": {
    "id": "uuid",
    "chat_room_id": "uuid",
    "sender_id": "uuid",
    "content": "Hello, world!",
    "created_at": "2025-10-17T11:00:00Z"
  }
}
```

#### message_updated

Triggered when a message is edited.

```json
{
  "type": "message_updated",
  "data": {
    "id": "uuid",
    "content": "Updated content",
    "is_edited": true,
    "edited_at": "2025-10-17T11:05:00Z"
  }
}
```

#### message_deleted

Triggered when a message is deleted.

```json
{
  "type": "message_deleted",
  "data": {
    "id": "uuid",
    "deleted_at": "2025-10-17T11:10:00Z"
  }
}
```

#### typing_start

User started typing.

```json
{
  "type": "typing_start",
  "data": {
    "chat_room_id": "uuid",
    "user_id": "uuid"
  }
}
```

#### typing_stop

User stopped typing.

```json
{
  "type": "typing_stop",
  "data": {
    "chat_room_id": "uuid",
    "user_id": "uuid"
  }
}
```

#### presence_updated

User presence changed.

```json
{
  "type": "presence_updated",
  "data": {
    "user_id": "uuid",
    "status": "online",
    "last_seen": "2025-10-17T11:00:00Z"
  }
}
```

#### reaction_added

Reaction added to message.

```json
{
  "type": "reaction_added",
  "data": {
    "message_id": "uuid",
    "user_id": "uuid",
    "emoji": "👍"
  }
}
```

#### participant_joined

User joined the room.

```json
{
  "type": "participant_joined",
  "data": {
    "chat_room_id": "uuid",
    "user_id": "uuid",
    "role": "member"
  }
}
```

#### participant_left

User left the room.

```json
{
  "type": "participant_left",
  "data": {
    "chat_room_id": "uuid",
    "user_id": "uuid"
  }
}
```

---

## Example Integration

### Complete Chat Flow

```javascript
// 1. Create a chat room
const createRoom = await fetch('http://localhost:9090/api/do', {
  method: 'POST',
  headers: {
    'Authorization': 'Bearer ' + jwtToken,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    action: 'chatRoomCreate',
    data: {
      name: 'Project Discussion',
      type: 'group',
      entity_type: 'project',
      entity_id: projectId
    }
  })
});

const room = await createRoom.json();
const roomId = room.data.id;

// 2. Add participants
await fetch('http://localhost:9090/api/do', {
  method: 'POST',
  headers: {
    'Authorization': 'Bearer ' + jwtToken,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    action: 'participantAdd',
    data: {
      chat_room_id: roomId,
      user_id: userId,
      role: 'member'
    }
  })
});

// 3. Connect WebSocket
const ws = new WebSocket('ws://localhost:9090/ws?jwt=' + jwtToken);

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  console.log('New event:', message.type, message.data);
};

// 4. Send a message
await fetch('http://localhost:9090/api/do', {
  method: 'POST',
  headers: {
    'Authorization': 'Bearer ' + jwtToken,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    action: 'messageSend',
    data: {
      chat_room_id: roomId,
      content: 'Hello, team!',
      type: 'text'
    }
  })
});

// 5. Start typing indicator
await fetch('http://localhost:9090/api/do', {
  method: 'POST',
  headers: {
    'Authorization': 'Bearer ' + jwtToken,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    action: 'typingStart',
    data: {
      chat_room_id: roomId
    }
  })
});

// 6. Load messages
const messages = await fetch('http://localhost:9090/api/do', {
  method: 'POST',
  headers: {
    'Authorization': 'Bearer ' + jwtToken,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    action: 'messageList',
    data: {
      chat_room_id: roomId,
      limit: 50,
      offset: 0
    }
  })
});

console.log('Messages:', await messages.json());
```

---

## Rate Limit Best Practices

1. **Implement exponential backoff** when receiving 429 errors
2. **Cache responses** when appropriate
3. **Batch requests** when possible
4. **Use WebSocket** for real-time updates instead of polling
5. **Monitor rate limit headers** to avoid hitting limits

---

## Support

- **Documentation**: https://docs.helixtrack.io
- **API Issues**: https://github.com/Helix-Track/Core/issues
- **Community**: https://community.helixtrack.io

---

**Last Updated**: 2025-10-17
**API Version**: 1.0.0

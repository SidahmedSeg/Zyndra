# Centrifugo Integration for Real-Time Logs

## Overview
This document outlines the changes needed to replace raw WebSocket implementation with Centrifugo for real-time log streaming in Click-to-Deploy.

## Benefits of Centrifugo

1. **Multiple Transport Support**: WebSocket, Server-Sent Events (SSE), HTTP-streaming
2. **Better Scalability**: Built-in horizontal scaling support
3. **Connection Management**: Automatic reconnection, presence tracking
4. **History & Presence**: Can store and replay recent messages
5. **Production Ready**: Battle-tested in production environments

## Required Changes to Specification

### 1. Technology Stack (Section 3.1)

**Current:**
```
WebSocket: nhooyr.io/websocket - Modern, RFC 6455 compliant
```

**Updated:**
```
Real-time Messaging: Centrifugo (centrifugo/centrifugo-go) - Production-ready real-time messaging server
```

### 2. Architecture Components (Section 2.2)

**Add to System Components:**
```
Component          Technology          Purpose
Centrifugo Server Centrifugo          Real-time messaging server for logs
Centrifugo Client  centrifugo-go       Go SDK for publishing messages
```

### 3. Project Structure (Section 11)

**Current:**
```
├── internal/
│   ├── stream/
│   │   └── logs.go              # WebSocket log streaming
```

**Updated:**
```
├── internal/
│   ├── stream/
│   │   ├── centrifugo.go        # Centrifugo client wrapper
│   │   └── logs.go              # Log publishing to Centrifugo
```

### 4. API Endpoints (Section 5.3.4)

**Current:**
```
WS  /deployments/:id/logs/stream  Stream logs in real-time
WS  /services/:id/logs/stream     Stream runtime logs in real-time
```

**Updated:**
```
GET /deployments/:id/logs/stream  Get Centrifugo connection token
GET /services/:id/logs/stream     Get Centrifugo connection token
```

**Note:** Clients connect directly to Centrifugo server using the token. The actual streaming happens via Centrifugo's WebSocket/SSE endpoints.

### 5. Configuration (Section 13.1)

**Add Environment Variables:**
```
Variable                Required  Description
CENTRIFUGO_URL          Yes       Centrifugo server URL (http://localhost:8000)
CENTRIFUGO_API_KEY      Yes       Centrifugo API key for publishing
CENTRIFUGO_SECRET       Yes       Centrifugo secret for JWT generation
CENTRIFUGO_HISTORY_TTL  No        Log history retention (default: 1h)
```

### 6. Implementation Details

#### 6.1 Centrifugo Client (internal/stream/centrifugo.go)

```go
package stream

import (
    "github.com/centrifugal/centrifugo/v5"
    "context"
)

type CentrifugoClient struct {
    client *centrifugo.Client
    secret string
}

func NewCentrifugoClient(url, apiKey, secret string) (*CentrifugoClient, error) {
    client, err := centrifugo.NewClient(centrifugo.Config{
        URL: url,
        APIKey: apiKey,
    })
    if err != nil {
        return nil, err
    }
    return &CentrifugoClient{
        client: client,
        secret: secret,
    }, nil
}

// PublishLog publishes a log message to a channel
func (c *CentrifugoClient) PublishLog(ctx context.Context, channel string, log LogEntry) error {
    data, _ := json.Marshal(log)
    _, err := c.client.Publish(ctx, channel, data)
    return err
}

// GenerateToken generates a JWT token for client connection
func (c *CentrifugoClient) GenerateToken(userID, channel string) (string, error) {
    claims := jwt.MapClaims{
        "sub": userID,
        "channels": []string{channel},
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(c.secret))
}
```

#### 6.2 Log Publishing (internal/stream/logs.go)

```go
package stream

import (
    "context"
    "time"
)

type LogEntry struct {
    Timestamp time.Time `json:"timestamp"`
    Level     string    `json:"level"` // info, warn, error
    Message   string    `json:"message"`
    Phase     string    `json:"phase,omitempty"` // clone, build, push, deploy
    Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// PublishDeploymentLog publishes a log entry to deployment channel
func (s *StreamService) PublishDeploymentLog(ctx context.Context, deploymentID string, entry LogEntry) error {
    channel := "deployment:" + deploymentID
    return s.centrifugo.PublishLog(ctx, channel, entry)
}

// PublishServiceLog publishes a log entry to service channel
func (s *StreamService) PublishServiceLog(ctx context.Context, serviceID string, entry LogEntry) error {
    channel := "service:" + serviceID
    return s.centrifugo.PublishLog(ctx, channel, entry)
}
```

#### 6.3 Worker Integration

In build/deploy workers, publish logs to Centrifugo instead of WebSocket:

```go
// During build process
stream.PublishDeploymentLog(ctx, deploymentID, LogEntry{
    Timestamp: time.Now(),
    Level:     "info",
    Message:   "Cloning repository...",
    Phase:     "clone",
})

stream.PublishDeploymentLog(ctx, deploymentID, LogEntry{
    Timestamp: time.Now(),
    Level:     "info",
    Message:   "Building container image...",
    Phase:     "build",
})
```

### 7. Frontend Changes (Section 3.3)

**Current:**
```javascript
// Native WebSocket API
const ws = new WebSocket('wss://api.projects.armonika.cloud/deployments/123/logs/stream');
```

**Updated:**
```javascript
// Using Centrifugo JavaScript SDK
import { Centrifuge } from 'centrifuge';

// Get token from API
const tokenResponse = await fetch('/api/deployments/123/logs/stream');
const { token } = await tokenResponse.json();

// Connect to Centrifugo
const centrifuge = new Centrifuge('wss://centrifugo.projects.armonika.cloud/connection/websocket', {
    token: token
});

const sub = centrifuge.newSubscription('deployment:123');
sub.on('publication', (ctx) => {
    const logEntry = ctx.data;
    // Display log entry in UI
    appendLog(logEntry);
});

sub.subscribe();
centrifuge.connect();
```

### 8. Centrifugo Configuration

**centrifugo.json:**
```json
{
  "token_hmac_secret_key": "your-secret-key",
  "admin_password": "admin-password",
  "admin_secret": "admin-secret",
  "allowed_origins": ["https://projects.armonika.cloud"],
  "history_size": 1000,
  "history_ttl": "1h",
  "presence": true,
  "join_leave": false,
  "websocket_compression": true
}
```

### 9. Deployment Architecture

**Updated Architecture Diagram:**
```
┌─────────────────────────────────────────────────────────────┐
│                    User Browser                             │
└───────────────────────┬─────────────────────────────────────┘
                        │
        ┌───────────────┴───────────────┐
        │                               │
        ▼                               ▼
┌───────────────┐           ┌──────────────────────┐
│ Click to      │           │   Centrifugo Server  │
│ Deploy API    │──────────▶│  (Real-time logs)    │
│               │  Publish  │                      │
└───────────────┘           └──────────────────────┘
        │                               │
        │                               │ Subscribe
        │                               ▼
        │                    ┌──────────────────────┐
        │                    │   User Browser      │
        │                    │  (WebSocket/SSE)    │
        │                    └──────────────────────┘
        │
        ▼
┌───────────────┐
│ OpenStack     │
│ Service       │
└───────────────┘
```

### 10. Migration Benefits Summary

1. **Better Client Support**: Works with WebSocket, SSE, or HTTP-streaming
2. **Scalability**: Can run multiple Centrifugo instances behind load balancer
3. **History**: Can replay recent logs if client reconnects
4. **Presence**: Know which users are viewing logs
5. **Production Ready**: Used by many production systems
6. **Easier Testing**: Can mock Centrifugo for testing

### 11. Dependencies Update

**go.mod additions:**
```
require (
    github.com/centrifugal/centrifugo/v5 v5.x.x
    github.com/golang-jwt/jwt/v5 v5.x.x
)
```

## Implementation Steps

1. **Install Centrifugo**: Deploy Centrifugo server alongside Click-to-Deploy
2. **Update Go Dependencies**: Add centrifugo-go SDK
3. **Replace WebSocket Handler**: Remove WebSocket endpoints, add token generation
4. **Update Workers**: Publish logs to Centrifugo channels
5. **Update Frontend**: Use Centrifugo JavaScript SDK
6. **Update Configuration**: Add Centrifugo environment variables
7. **Testing**: Test reconnection, history, and multiple clients

## Backward Compatibility

- Old WebSocket endpoints can be deprecated gradually
- Both systems can run in parallel during migration
- Frontend can detect Centrifugo availability and fallback to WebSocket if needed


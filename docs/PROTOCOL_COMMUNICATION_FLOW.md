# Protocol Communication Flow Verification

**Date**: 2025-11-15
**Status**: ✅ Verified - Frontend and Backend are aligned

---

## Summary

This document verifies that the frontend and backend are using consistent envelope-based protocol for all message types.

---

## Client → Server Communication

### PlayerInput Flow

#### Frontend Sending (TypeScript)

**File**: `frontend/src/network/client.ts:387-393`

```typescript
sendPlayerInput(input: PlayerInput): void {
  if (!this.isConnected()) return;

  // Wrap player input in envelope as per protocol specification
  const message = createPlayerInputMessage(input);
  this.sendMessage(message);
}
```

**Message Creation**: `frontend/src/network/protocols.ts:111-116`

```typescript
export function createPlayerInputMessage(input: PlayerInput): RequestEnvelope {
  return {
    envelope_type: REQUEST_TYPES.PLAYER_INPUT,  // "player_input"
    payload: input
  };
}
```

**Wire Format**:
```json
{
  "envelope_type": "player_input",
  "payload": {
    "MoveUp": true,
    "MoveDown": false,
    "MoveLeft": false,
    "MoveRight": true,
    "RotateLeft": false,
    "RotateRight": false,
    "SwitchWeapon": false,
    "Reload": false,
    "FastReload": false,
    "Fire": false,
    "Timestamp": 1731686400000
  }
}
```

#### Backend Receiving (Go)

**File**: `internal/app/client.go:183-227`

```go
func (c *websocketClient) readPump() {
	for {
		var msg protocol.RequestEnvelope  // ✓ Expects RequestEnvelope

		data, err := c.conn.ReadMessage()
		// ...

		if errDecode := c.codec.Decode(data, &msg); errDecode != nil {
			// Error handling
		}

		parsed, err := protocol.GetPayloadStruct(msg.EnvelopeType)  // ✓ Uses EnvelopeType
		// ...

		command := protocol.RequestCommand{
			ClientID:      c.id,
			EnvelopeType:  msg.EnvelopeType,  // ✓ Passes envelope type
			Payload:       msg.Payload,
			ParsedPayload: parsed,
			ReceivedTime:  utils.Now(),
		}
		// Routes to subscriptions
	}
}
```

**Protocol Definition**: `internal/protocol/protocol.go:28-31`

```go
type RequestEnvelope struct {
	EnvelopeType RequestEnvelopeType `json:"envelope_type"`  // ✓ Matches frontend
	Payload      json.RawMessage     `json:"payload"`
}
```

**Verification**: ✅ **ALIGNED**
- Frontend sends: `{envelope_type, payload}`
- Backend expects: `{envelope_type, payload}`
- Field names match exactly

---

## Server → Client Communication

### GameUpdate Flow

#### Backend Sending (Go)

**File**: `internal/game/room.go:231-249`

```go
func (r *Room) broadcastGameUpdate() {
	gameUpdate := r.state.ToClientState()

	payloadBytes, err := json.Marshal(gameUpdate)
	// ...

	envelope := protocol.ResponseEnvelope{
		EnvelopeType: protocol.GameUpdateEnvelope,  // ✓ Uses EnvelopeType field
		Payload:      json.RawMessage(payloadBytes),
	}

	r.outgoing <- UpdateMessage{
		ToSessions: r.players.AllSessionIDs(),
		Envelope:   envelope,  // ✓ Sends ResponseEnvelope
	}
}
```

**Protocol Definition**: `internal/protocol/protocol.go:33-36`

```go
type ResponseEnvelope struct {
	EnvelopeType ResponseEnvelopeType `json:"envelope_type"`  // ✓ JSON tag matches frontend
	Payload      json.RawMessage      `json:"payload"`
}
```

**Wire Format**:
```json
{
  "envelope_type": "game_update",
  "payload": {
    "players": {...},
    "walls": [...],
    "projectiles": [...],
    "timestamp": 1731686400000
  }
}
```

#### Frontend Receiving (TypeScript)

**File**: `frontend/src/network/client.ts:211-258`

```typescript
private handleServerMessage(envelope: ResponseEnvelope): void {
  console.log('Received server message:', envelope.envelope_type);  // ✓ Uses envelope_type

  switch (envelope.envelope_type) {  // ✓ Switches on envelope_type
    case RESPONSE_TYPES.GAME_UPDATE:
      this.handleGameUpdate(envelope.payload);  // ✓ Extracts payload
      break;
    // ... other cases
  }
}
```

**Protocol Definition**: `frontend/src/network/protocols.ts:15-18`

```typescript
export interface ResponseEnvelope {
  envelope_type: string;  // ✓ Matches backend JSON tag
  payload: any;
}
```

**Verification**: ✅ **ALIGNED**
- Backend sends: `{envelope_type, payload}` (Go struct with json tags)
- Frontend expects: `{envelope_type, payload}` (TypeScript interface)
- Field names match exactly via JSON serialization

---

## All Message Types Verification

### Client → Server

| Message Type | Frontend Creates | Backend Receives | Status |
|--------------|------------------|------------------|--------|
| `player_input` | `createPlayerInputMessage()` | `readPump()` → `RequestEnvelope` | ✅ Aligned |
| `room_list` | `createRoomListRequest()` | `readPump()` → `RequestEnvelope` | ✅ Aligned |
| `join_room` | `createJoinRoomRequest()` | `readPump()` → `RequestEnvelope` | ✅ Aligned |
| `leave_room` | `createLeaveRoomRequest()` | `readPump()` → `RequestEnvelope` | ✅ Aligned |

**All use**: `RequestEnvelope{envelope_type, payload}`

### Server → Client

| Message Type | Backend Sends | Frontend Receives | Status |
|--------------|---------------|-------------------|--------|
| `game_update` | `broadcastGameUpdate()` → `ResponseEnvelope` | `handleServerMessage()` | ✅ Aligned |
| `static_data` | `SendStaticData()` → `ResponseEnvelope` | `handleServerMessage()` | ✅ Aligned |
| `system_set_session` | `client.Send()` → `ResponseEnvelope` | `handleServerMessage()` | ✅ Aligned |
| `system_notify` | `client.Send()` → `ResponseEnvelope` | `handleServerMessage()` | ✅ Aligned |
| `error_invalid_session` | `sendSessionInvalidMessage()` → `ResponseEnvelope` | `handleServerMessage()` | ✅ Aligned |

**All use**: `ResponseEnvelope{envelope_type, payload}`

---

## JSON Field Mapping

### Go Struct Tags → JSON → TypeScript

**Go (Backend)**:
```go
type RequestEnvelope struct {
    EnvelopeType RequestEnvelopeType `json:"envelope_type"`
    Payload      json.RawMessage     `json:"payload"`
}

type ResponseEnvelope struct {
    EnvelopeType ResponseEnvelopeType `json:"envelope_type"`
    Payload      json.RawMessage      `json:"payload"`
}
```

**JSON (Wire Format)**:
```json
{
  "envelope_type": "...",
  "payload": {...}
}
```

**TypeScript (Frontend)**:
```typescript
interface RequestEnvelope {
  envelope_type: string;
  payload: any;
}

interface ResponseEnvelope {
  envelope_type: string;
  payload: any;
}
```

**Mapping**: ✅ **PERFECT MATCH**
- Go `json:"envelope_type"` → JSON `"envelope_type"` → TypeScript `envelope_type`
- Go `json:"payload"` → JSON `"payload"` → TypeScript `payload`

---

## Serialization Flow

### Client → Server

```
TypeScript Object                     JSON String                          Go Struct
─────────────────                     ───────────                          ─────────
RequestEnvelope{                  →   {"envelope_type":"player_input",  →  RequestEnvelope{
  envelope_type: "player_input",      "payload":{...}}                      EnvelopeType: PlayerInputEnvelope,
  payload: {...}                                                            Payload: json.RawMessage
}                                                                          }
│                                     │                                    │
└─ JSON.stringify()                   └─ WebSocket                         └─ json.Unmarshal()
```

### Server → Client

```
Go Struct                            JSON String                          TypeScript Object
─────────                            ───────────                          ─────────────────
ResponseEnvelope{                →   {"envelope_type":"game_update",   →  ResponseEnvelope{
  EnvelopeType: GameUpdateEnvelope,   "payload":{...}}                     envelope_type: "game_update",
  Payload: json.RawMessage                                                 payload: {...}
}                                                                         }
│                                    │                                    │
└─ json.Marshal()                    └─ WebSocket                         └─ JSON.parse()
```

---

## Critical Points Verified

### 1. ✅ Envelope Structure Consistency
- Both directions use envelope wrapper
- Field names match via JSON tags
- No raw message sending

### 2. ✅ Field Naming Convention
- Backend: `EnvelopeType` (PascalCase struct field)
- Backend JSON tag: `envelope_type` (snake_case)
- Frontend: `envelope_type` (snake_case)
- **Result**: Perfect match in JSON

### 3. ✅ Message Type Constants
- Backend: `protocol.PlayerInputEnvelope = "player_input"`
- Frontend: `REQUEST_TYPES.PLAYER_INPUT = "player_input"`
- **Result**: String values match

### 4. ✅ Payload Handling
- Backend: `json.RawMessage` (deferred parsing)
- Frontend: `any` type (flexible parsing)
- **Result**: Both support arbitrary JSON payloads

---

## Test Cases

### Test 1: PlayerInput Message

**Frontend Sends**:
```json
{
  "envelope_type": "player_input",
  "payload": {
    "MoveUp": true,
    "MoveDown": false,
    "MoveLeft": false,
    "MoveRight": true,
    "RotateLeft": false,
    "RotateRight": false,
    "SwitchWeapon": false,
    "Reload": false,
    "FastReload": false,
    "Fire": false,
    "Timestamp": 1731686400000
  }
}
```

**Backend Receives**:
```go
RequestEnvelope{
    EnvelopeType: "player_input",
    Payload: json.RawMessage(`{"MoveUp":true,"MoveDown":false,...}`)
}
```

**Status**: ✅ Will parse correctly

### Test 2: GameUpdate Message

**Backend Sends**:
```json
{
  "envelope_type": "game_update",
  "payload": {
    "players": {...},
    "walls": [...],
    "projectiles": [...],
    "timestamp": 1731686400000
  }
}
```

**Frontend Receives**:
```typescript
{
  envelope_type: "game_update",
  payload: {
    players: {...},
    walls: [...],
    projectiles: [...],
    timestamp: 1731686400000
  }
}
```

**Status**: ✅ Will parse correctly

---

## Conclusion

### ✅ Protocol Communication Status: **FULLY ALIGNED**

**All systems verified**:
1. ✅ Frontend sends envelopes with correct field names
2. ✅ Backend receives envelopes with matching structure
3. ✅ Backend sends envelopes with correct field names
4. ✅ Frontend receives envelopes with matching structure
5. ✅ JSON serialization/deserialization is consistent
6. ✅ All message types use the same envelope pattern

**No communication issues expected.**

### Recent Fixes Applied

1. **Envelope Naming Standardization** (commit: cb01ec7)
   - Unified `EnvelopeType` struct field names
   - Unified `envelope_type` JSON field names

2. **PlayerInput Envelope Wrapping** (commit: 072ad20)
   - Changed frontend from raw JSON to envelope format
   - Now matches backend expectations

3. **Projectile Field Alignment** (current)
   - Changed frontend `Velocity` → `Direction`
   - Added missing fields: `Speed`, `Range`, `Damage`

### Files Verified

**Frontend**:
- ✅ `frontend/src/network/client.ts`
- ✅ `frontend/src/network/protocols.ts`
- ✅ `frontend/src/state.ts`

**Backend**:
- ✅ `internal/app/client.go`
- ✅ `internal/game/room.go`
- ✅ `internal/protocol/protocol.go`

**Documentation**:
- ✅ `docs/PROTOCOL_SPECIFICATION.md`

---

**Last Updated**: 2025-11-15
**Protocol Version**: 1.1

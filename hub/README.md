# wshub

A WebSocket gateway for backends that have no native WebSocket support (PHP, for example). The backend talks to the
gateway over a plain HTTP API, and browsers connect to the gateway over WebSocket.

## Envelope

Every message on the WebSocket, in both directions, is a JSON envelope:

```json
{"id": "a1b2c3", "type": "some_type", "data": {"any": "fields"}}
```

The HTTP API takes the same envelope as the request body, with one more field, `waitResponse`.

## Waiting for a reply from the WS client

A backend can send a message to a client and wait for the client's answer in the same HTTP request. This works for
`POST /api/client-message` and `POST /api/registered-client-message`. It does not work for `POST /api/tag-message`,
because one request would get many answers.

### 1. The backend sends the question

```
POST /api/registered-client-message?clientId=<client id>&messageId=<message id>
Content-Type: application/json

{"type": "correct_upload_file", "data": {"message": "The file already exists", "fileName": "a.txt"}, "waitResponse": 60000}
```

- `waitResponse` is the time to wait in milliseconds. `0` (or no field) means do not wait: the message is sent and the
  gateway answers `200` with an empty body.
- `messageId` is what ties the question to the answer. It may contain letters, digits, `-` and `_`, up to 64
  characters. A request that waits needs a non-empty `messageId` that no other waiting request is using at the same
  time. Otherwise the request gets `504` at once.

### 2. The WS client receives the question

The `id` of the envelope the client receives is the `messageId` of the request:

```json
{"id": "<message id>", "type": "correct_upload_file", "data": {"message": "The file already exists", "fileName": "a.txt", "message_type": "correct_upload_file", "message_id": "<message id>"}}
```

`data` holds the fields the backend sent, plus `message_type` and `message_id`, which the gateway adds.

### 3. The WS client answers

**The answer must be an envelope with the same `id`.** This is how the gateway matches the answer to the waiting
request:

```json
{"id": "<message id>", "type": "upload_file_user_response", "data": {"correctedFileName": "b.txt", "cancelUpload": false}}
```

- `type` can be anything. The gateway does not look at it.
- The answer must come from the same WS connection that received the question. An answer from another connection
  releases the waiting request with `504`.
- An answer that comes after the wait has ended, or with an `id` nobody waits for, is dropped and logged.

### 4. The backend gets the answer

The HTTP request ends with `200` and the answer as an envelope, with the `id` of the request:

```json
{"id": "<message id>", "type": "upload_file_user_response", "data": {"correctedFileName": "b.txt", "cancelUpload": false, "message_type": "upload_file_user_response", "message_id": "<message id>"}}
```

`data` holds the fields of the answer, plus `message_type` and `message_id`, which the gateway adds.

### Other results

| Status | Body | When |
|---|---|---|
| `200` | empty | `waitResponse` is `0`: the message was sent, nothing was waited for |
| `200` | the answer | the client answered in time |
| `400` | `{"error": "..."}` | invalid `clientId`, `messageId` or envelope |
| `404` | `{"error": "..."}` | no such client: for `/api/registered-client-message` the session was not registered, or has no WS connection; for `/api/client-message` there is no connection with that id |
| `504` | `{"error": "..."}` | no answer within `waitResponse`, the client disconnected while waiting, or the `messageId` was empty or already in use |

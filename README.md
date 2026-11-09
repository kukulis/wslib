# wslib

A WebSocket gateway for backends without native WebSocket support (PHP, for example).

- `hub/` is the gateway, written in Go. Browsers connect to it over WebSocket, backends talk to it over a plain HTTP API.
- `wshub_user/pkg/WsGatewayClient` is the PHP client of that HTTP API.
- The rest of `wshub_user/` is a demo application that uses both.

## Starting the project

You need Docker with the compose plugin.

1. Create the network once. The compose file expects it to exist already:

       docker network create wslib-network

2. Start the containers from the repo root:

       docker compose up -d

   | Container | What it does |
   |---|---|
   | `wshub` | the gateway (`go run ./cmd/wshub`), listens on port 8585 inside the network |
   | `wshub_user` | PHP-FPM with the demo code and the PHP client |
   | `wshub_user_web` | nginx, serves the demo on port 8787 and proxies `/ws` to the gateway |

3. Install the PHP dependencies, if `wshub_user/vendor` is missing:

       docker compose exec -u 1000:1000 wshub_user bash -c "cd wshub_user && composer install"

4. Open http://localhost:8787/ . The page lists the examples.

The gateway has no published port: browsers reach it through nginx (`/ws`), and PHP reaches it as `http://wshub:8585`.
Note: the map examples load map tiles from the public OpenStreetMap server, so they need internet access.

### Tests

    docker compose exec -u 1000:1000 wshub_user bash -c "cd wshub_user && vendor/bin/phpunit tests"
    docker compose exec wshub bash -c "cd /app && go test ./..."

## Examples

All of them follow the same pattern. The PHP page registers a session with the gateway (a client id, a token and
optional tags), then gives the id and token to the browser. The browser opens a WebSocket and logs in with them.
From then on the backend can reach that browser by its client id, or every browser with a tag.

### 1. Delayed confirmation: `public/confirm2.php`

The browser starts a long background job over plain HTTP. The job finishes later, outside of the request, and the
backend notifies the page over WebSocket. Nothing polls.

Open the page and press the button. The result appears after a random delay of up to 20 seconds.

### 2. Moving objects on a map: `public/map4.php`

The backend pushes the positions of several moving objects in one `moving_objects` message, and the page moves a
marker for each of them on a Leaflet map. All pages register with the tag `map`, so one message to the tag moves the
objects on every open page. (`map3.php` is the same example with a single object.)

1. Open http://localhost:8787/map4.php (open it in two tabs, if you like).
2. Start the movement in the `wshub_user` container:

       docker compose exec -u 1000:1000 wshub_user bash -c "cd wshub_user && php long_running/imitate_movement_registered.php map"

### 3. File upload with a question to the user: `public/upload.php`

The user uploads a file. If the name is not allowed or the file already exists, the backend asks the user for
another name, **waits for the answer in the same PHP call**, and then continues. The code that shows it is
`FileUploadHandler::uploadFileCycle` in `wshub_user/src/Handler`. The user can correct the name or cancel.

## Contracts

### The envelope

Every message on the WebSocket, in both directions, is a JSON envelope:

```json
{"id": "a1b2c3", "type": "some_type", "data": {"any": "fields"}}
```

- `id` ties a question to its answer.
- `type` tells the receiver what the message is. The receiver reads the envelope first, picks the object for that
  type, and only then reads `data` into it.
- `data` is the payload, free-form.

The HTTP API takes the same envelope as the request body, with one more field, `waitResponse` (milliseconds).
The reply of the WS client comes back as an envelope too. See [`hub/README.md`](hub/README.md) for the details of
waiting for a reply.

### The HTTP API

| Endpoint | Purpose |
|---|---|
| `POST /api/register-session?clientId=&sessionToken=&tags=a,b` | the backend announces a client that is about to log in |
| `POST /api/registered-client-message?clientId=&messageId=` | send a message to one registered client; may wait for the answer |
| `POST /api/tag-message?tag=&messageId=` | send a message to every logged in client with the tag; cannot wait |
| `POST /api/client-message?clientId=&messageId=` | send a message to a connection by its connection id; may wait for the answer |

### The login of a browser

The first message of a WebSocket connection must be the login, with the id and token the backend registered:

```json
{"id": "login", "type": "login", "data": {"clientId": "...", "token": "..."}}
```

The gateway answers with `login_result`. On success, `data.token` is a new token: the used one is spent, so a
reconnect must use the new one.

The PHP client wraps the API: `WsGatewayClient\Client` takes an `Envelope` and returns a `SendMessageResponse`
(`responseCode`, `errorMessage` and the reply as `wsMessage`, an `Envelope`).

## Development

Run docker commands as user 1000:1000 where possible:

    docker exec -itu 1000:1000 wshub bash

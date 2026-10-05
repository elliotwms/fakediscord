# fakediscord

The aim of `fakediscord` is to replicate the behaviour of the Discord HTTP and Websocket APIs, based on the documentation and observed behaviour, in order to enable the integration testing of Discord bots without calling the real Discord API. 

Analogous to [LocalStack](https://github.com/localstack/localstack), `fakediscord` should be run locally using Docker when running your bot's tests.

While written in Go, `fakediscord` can be used to test bots in any language, provided they adhere to Discord's specifications.

```mermaid
flowchart LR
t["Your tests"] --> b["Your Bot"] --> d["Discord"]
b --> f["fakediscord"]
t --> f["fakediscord"]
```


`fakediscord` fakes the HTTP and WebSocket endpoints of the Discord API, triggering corresponding events via the WebSocket connection. `fakediscord` pairs well with (and is based on the hard work of) [bwmarrin/discordgo](https://github.com/bwmarrin/discordgo). 

Of course, you should also test your bot manually before releasing to the public: there are many features currently not present, such as authorization, -- any action is currently allowed.

## Usage

`fakediscord` should work with any Discord client in any language, and is intended to be run via Docker:

```shell
docker run -p 8080:8080 ghcr.io/elliotwms/fakediscord:{version}
```

It is possible to provide a `config.yml` file to bootstrap users and guilds (todo: document config): 

```yaml
services:
  fakediscord:
    image: ghcr.io/elliotwms/fakediscord:{version}
    ports:
      - 8080:8080
    volumes:
      - ${PWD}/fakediscord.yaml:/config.yml:ro
```

The following environment variables can be set:

| Variable      | Default      | Description                                      |
|---------------|--------------|--------------------------------------------------|
| `PORT`        | `8080`       | Port to listen on                                |
| `CONFIG_PATH` | `config.yml` | Path to the config file. A missing file is fine |

The gateway URL returned by `GET /gateway` uses the host the request was made to, so `fakediscord` can be reached by a docker-compose service name (e.g. `http://fakediscord:8080/`) or a remapped port.

`fakediscord` provides a Go client as a convenience wrapper for internal endpoints, as well as helpers to point discordgo (or any `http.Client`) at `fakediscord`, which can be found in `pkg/fakediscord`.

Point your session at `fakediscord` before opening it, then use it as normal:

```go
package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/elliotwms/fakediscord/pkg/fakediscord"
)

const baseURL = "http://localhost:8080/"

func main() {
	session, _ := discordgo.New("Bot your-bot-token")

	// send the session's requests (and so its gateway connection) to fakediscord
	if err := fakediscord.ConfigureSession(session, baseURL); err != nil {
		panic(err)
	}

	// Client for internal endpoints (e.g. interactions)
	c := fakediscord.NewClient("your-bot-token").WithBaseURL(baseURL)
}
```

`ConfigureSession` wraps the session's HTTP client with `fakediscord.Transport`, which sends any request addressed to `discord.com` to `fakediscord` instead. If your bot calls the Discord API without discordgo, you can use `Transport` with your own `http.Client`:

```go
u, _ := url.Parse("http://localhost:8080/")
client := &http.Client{Transport: fakediscord.Transport(u, nil)}
```

`fakediscord.Configure(baseURL)`, which overrides discordgo's package-level endpoints, is deprecated. It only covers some endpoints, and applies to every session in the process.

### Authentication

* Any token value will pass authentication (`Bot {token}`). A missing `Authorization` header returns `401`
* If the token matches one specified in the config then the relevant user will be authenticated
* Otherwise, a user will be generated with the token value as the username
* For testing purposes, all users are assumed to be in all guilds

### Interactions

`fakediscord` provides an endpoint for triggering interactions, which would normally only be possible via a user initiating via the UI. A `POST` of an `InteractionCreate` event to `/api/:version/interactions` will create an interaction. If the interaction has no `member` or `user`, the authenticated user is set as the invoking member.

A suggested pattern for testing interactions within a webhook application would be as follows: 

1. Build the expected interaction within your test suite
2. Create the initial interaction in `fakediscord`. This will provide you with IDs, tokens etc
3. Send the interaction to your application's endpoint
4. Your application will likely call the [interaction's callback url](https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-callback) to acknowledge the interaction

```mermaid
sequenceDiagram
    participant t as Tests
    participant a as App
    participant d as fakediscord
    
    t->>t: Set up interaction
    t->>d: POST /interactions
    d->>t: 201 Created: Interaction
    t->>a: Interaction
    activate a
    a->>d: POST /interactions/:id/:token/callback
    d->>a: 204 No Content
    a->>a: Process interaction
    a->>d: POST/webhooks/:appID/:token/@original
    d->>a: 200 OK
    a->>t: 202 Accepted
    deactivate a
    t->>d: GET /webhooks/:appID/:token/@original
    d->>t: 200 OK: Message
    t->>t: Assert on message
```

## Features

`fakediscord` currently supports the following API operations, and emits the corresponding [events](https://discord.com/developers/docs/topics/gateway-events):

#### Gateway

* Get Gateway
* Get Gateway Bot
* Connect
  * `HELLO`
  * `READY`
  * Heartbeats (resuming is not supported)
  * [`GUILD_CREATE`](https://discord.com/developers/docs/events/gateway-events#guild-create)

### Guilds

* [Create](https://discord.com/developers/docs/resources/guild#create-guild)
  * [`GUILD_CREATE`](https://discord.com/developers/docs/events/gateway-events#guild-create)
* [Get](https://discord.com/developers/docs/resources/guild#get-guild)
* [Delete](https://discord.com/developers/docs/resources/guild#delete-guild)
  * [`GUILD_DELETE`](https://discord.com/developers/docs/events/gateway-events#guild-delete)
* [Get channels](https://discord.com/developers/docs/resources/guild#get-guild-channels)
* [Create channel](https://discord.com/developers/docs/resources/guild#create-guild-channel)
  * [`CHANNEL_CREATE`](https://discord.com/developers/docs/events/gateway-events#channel-create)

### Channels

* [Get](https://discord.com/developers/docs/resources/channel#get-channel)
* [Delete](https://discord.com/developers/docs/resources/channel#deleteclose-channel)
  * [`CHANNEL_DELETE`](https://discord.com/developers/docs/events/gateway-events#channel-delete)
* [Get Pinned Messages](https://discord.com/developers/docs/resources/channel#get-pinned-messages)
* [Pin Message](https://discord.com/developers/docs/resources/channel#pin-message)
  * [`CHANNEL_PINS_UPDATE`](https://discord.com/developers/docs/events/gateway-events#channel-pins-update)
* [Unpin Message](https://discord.com/developers/docs/resources/channel#unpin-message)
  * [`CHANNEL_PINS_UPDATE`](https://discord.com/developers/docs/events/gateway-events#channel-pins-update)

### Messages

* [Create Message](https://discord.com/developers/docs/resources/message#create-message)
  * [`MESSAGE_CREATE`](https://discord.com/developers/docs/events/gateway-events#message-create)
  * Supports basic, embeds and multipart 
* [Get Message](https://discord.com/developers/docs/resources/message#get-channel-message)
* [Delete Message](https://discord.com/developers/docs/resources/message#delete-message)
  * [`MESSAGE_DELETE`](https://discord.com/developers/docs/resources/message#delete-message)
* [Get Message Reactions](https://discord.com/developers/docs/resources/message#get-reactions)
* [Create Reaction](https://discord.com/developers/docs/resources/message#create-reaction)
  * [`MESSAGE_REACTION_ADD`](https://discord.com/developers/docs/events/gateway-events#message-reaction-add)
* [Delete Own/User Reaction](https://discord.com/developers/docs/resources/message#delete-user-reaction)
  * [`MESSAGE_REACTION_REMOVE`](https://discord.com/developers/docs/events/gateway-events#message-reaction-remove)
* [Delete Reactions](https://discord.com/developers/docs/resources/message#delete-all-reactions)
  * [`MESSAGE_REACTION_REMOVE_ALL`](https://discord.com/developers/docs/events/gateway-events#message-reaction-remove-all)

### Interactions

* Create (see [docs](#interactions))
* [Callback](https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-callback)
* [Get Original Response](https://discord.com/developers/docs/interactions/receiving-and-responding#get-original-interaction-response)
* [Edit Original Response](https://discord.com/developers/docs/interactions/receiving-and-responding#edit-original-interaction-response)

## Examples

Check out how the following projects use `fakediscord` for inspiration:

### [Pinbot](https://github.com/elliotwms/pinbot/tree/master/tests)

* Docker [Compose](https://github.com/elliotwms/pinbot/blob/master/compose.yaml) contains Pinbot config, including the bot user in [fakediscord.yaml](https://github.com/elliotwms/pinbot/blob/master/fakediscord.yaml)
* [TestMain](https://github.com/elliotwms/pinbot/blob/20debf13a3dff8e58b7d61ec5e04c18c1542be3d/tests/setup_test.go#L21) calls `fakediscord.Configure` (now deprecated in favour of `ConfigureSession`) to set base URLs etc, sets up the client, creates a test guild for the run and opens a general session for the test suite
* Individual tests then create channels in the test guild to execute their tests within ([example](https://github.com/elliotwms/pinbot/blob/20debf13a3dff8e58b7d61ec5e04c18c1542be3d/tests/pin_test.go#L7))

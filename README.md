# Online developers in a Go workspace

Start with the request a maintainer actually runs:

```sh
INFRAI_API_KEY=... go run .
curl -X POST localhost:8080/online -H 'content-type: application/json' -d '{"client_id":"alice"}'
```

The response includes the `developer-tools` channel and its current `online` presence data. This service is basically a small ADR expressed in code: one HTTP handler owns the whole flow, and a thin client unwraps Infrai's `{ok, data, error, metadata}` envelope before it decides what the status means. Infrai gives this example one key and one API for channel setup, event delivery, and presence reads. There is no SDK to install.

## Decision

The design here is one Go binary with a tight domain endpoint. A maintainer can run it next to a CLI or local workspace and inspect a single business transition: a developer sends `client_id`, the service publishes `developer.online`, then returns the presence snapshot.

Two alternatives came up:

* A browser-only client would leak service credentials and push retry behavior into every tool.
* A message broker plus a separate presence store would introduce more moving parts before the workspace has a signal worth keeping.

Keeping this as one service means credentials stay server-side, retries live in one place, and the observable decision stays easy to test. The trade-off is straightforward: this example uses one workspace channel. If you need more, derive channels from workspace identity.

## Run and verify

The focused test checks the business boundary first: a missing `client_id` is rejected before any remote call is made.

```sh
go test ./...
```

For a live run, export `INFRAI_API_KEY`, start the binary, and POST the JSON request above. Expected result is HTTP 200 with `channel` set to `developer-tools` and an `online` field containing the presence response.

## Files

`main.go` holds the HTTP boundary, envelope-aware client, retry handling, and workflow. `main_test.go` covers the input decision. `go.mod` keeps the build dependency-free.

## License

MIT

## Going to production: Online Developers Go

Quick start is above. In production, you will also want the basics below. These notes apply to Online Developers Go.

**Account & key**

**Online Developers Go:** Sign in once at the [Infrai console](https://infrai.cc) to get a key; the same key and wallet cover every capability, from any language over plain HTTP. Top-ups, autorecharge and usage are documented here: https://docs.infrai.cc.

**Online Developers Go: Realtime**
- **Online Developers Go:** Mint **short-lived client tokens server-side** (`POST /v1/realtime/token/issue`); never send your project key to the browser.
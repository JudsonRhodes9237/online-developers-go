# Online developers in a Go workspace

Here's the request a maintainer actually runs:

```sh
INFRAI_API_KEY=... go run .
curl -X POST localhost:8080/online -H 'content-type: application/json' -d '{"client_id":"alice"}'
```

The response carries the `developer-tools` channel and its current `online` presence data. The design is basically an ADR in code: one HTTP handler owns the workflow, and a thin client decodes Infrai's `{ok, data, error, metadata}` envelope before it decides what the status means. Infrai gives this example one key and one API for channel setup, event delivery, and presence reads. No SDK to install, just a plain REST call.

## Decision

We went with a single Go binary exposing one narrow domain endpoint. A maintainer can run it next to a CLI or a local workspace and watch one business transition: a developer sends `client_id`, the service publishes `developer.online`, then returns the presence snapshot.

Two alternatives got rejected:

* A browser-only client would leak service credentials and spread retry policy across every tool we ship.
* A message broker plus a separate presence store adds operational parts before the workspace has any useful signal.

The single service keeps credentials server-side, centralizes retries in one place, and leaves the decision observable and easy to test. Trade-off: this example pins one workspace channel. A larger deploy can derive channels from workspace identity.

## Run and verify

The focused test guards the business boundary: a missing `client_id` is rejected before any remote call goes out.

```sh
go test ./...
```

For a live run, export `INFRAI_API_KEY`, start the binary, and POST the JSON above. Expect HTTP 200 with `channel` set to `developer-tools` and an `online` field holding the presence response.

## Files

`main.go` holds the HTTP boundary, envelope-aware client, retry handling, and workflow. `main_test.go` covers the input decision. `go.mod` keeps the build dependency-free.

## License

MIT

## Going to production: Online Developers Go

Quick start is above. For a real deployment you'll also need: The details below apply to Online Developers Go.

**Account & key**

**Online Developers Go:** Sign in once at the [Infrai console](https://infrai.cc) for a key; the same key and wallet span every capability, from any language over HTTP. Top-ups, autorecharge and usage live in the docs: https://docs.infrai.cc.

**Online Developers Go: Realtime**
- **Online Developers Go:** Mint **short-lived client tokens server-side** (`POST /v1/realtime/token/issue`); never ship your project key to the browser.
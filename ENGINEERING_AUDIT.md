# Engineering Audit — atendi9-meta (`github.com/atendi9/meta`)

> Read-only audit. Branch audited: `fix/critical-bugs-and-full-coverage`.
> Verified with `go build ./...` (clean), `go vet ./...` (clean),
> `go test ./... -cover` (all pass; coverage: root 100%, uuid 96.0%,
> whatsapp 99.6%, xhttp 100%, xjson 100%).

## Summary
- critical: 0 · high: 3 · medium: 7 · low: 8
- Top risks:
  - `RequestError.Error()` / `.JSON()` expose the raw Meta error payload to
    callers and logs; Meta error bodies can echo request context and headers,
    so a leaked `Bearer` token is plausible if callers log the error verbatim.
  - HTTP response bodies are not closed on several paths (`WebhookSubscribe`,
    `TmplAnalytics.Enable`) — connection/socket leak under load.
  - API-level errors are silently swallowed in the calling flow
    (`AcceptCall`, `WebhookPreAccept`) — failures look like successes.

## Findings

### High

- [error-handling/resource] `whatsapp/webhook_subscribe.go:35-52` —
  `WebhookSubscribe` performs two `whats.Post(...)` calls and never closes the
  returned `res.Body` on the success path. `xhttp.Client.executeRequest`
  returns a live `*http.Response` whose body must be drained+closed to allow
  HTTP keep-alive connection reuse. Under repeated webhook subscription this
  leaks sockets/file descriptors.
  Fix: capture the response and `defer res.Body.Close()` for each POST (and
  ideally `io.Copy(io.Discard, res.Body)` before closing).

- [error-handling/resource] `whatsapp/template_analytics.go:412-425` —
  `TmplAnalytics.Enable` discards the response (`_, err := whats.Post(...)`)
  and therefore never closes its body — same FD/connection leak. The function
  also ends with `return err == nil` after already handling `err != nil`
  above, which is dead/confusing logic.
  Fix: bind the response, `defer res.Body.Close()`, and simplify the return.

- [error-handling] `whatsapp/calling.go:49-71,141-183` — `AcceptCall` returns
  the call `id` even when the POST failed (only logs), and `WebhookPreAccept`
  returns no error to the caller at all. A failed `accept`/`pre_accept`
  request is indistinguishable from success, so a dropped call is never
  surfaced. `AcceptCall` also does not inspect the API JSON body for an
  application-level error.
  Fix: return `error` from `AcceptCall`/`WebhookPreAccept` (or an explicit
  status) so callers can react; close `res.Body` (currently done, ok).

### Medium

- [security] `xhttp/error.go:29-41` & `whatsapp/client.go:83-88` — Meta Graph
  API error responses are returned verbatim by `RequestError.Error()`. The
  `Authorization: Bearer <token>` header is attached to every request
  (`Client.Headers`). If the backend logs `err.Error()` (and the Atendi9-API
  is known to surface raw Meta errors), there is a realistic path for token
  or PII exposure. No redaction layer exists.
  Fix: add a redaction step that strips `access_token`, `Authorization`, and
  known token-shaped fields before the payload is stringified, or expose only
  a sanitized `code`/`message` subset.

- [correctness] `xhttp/client.go:134-139` — when the status is `>= 400`,
  `executeRequest` returns *both* a non-nil `*http.Response` and a non-nil
  `*RequestError`. The error's `Payload` is `res.Body`; callers that do
  `defer res.Body.Close()` AND read `err.JSON()` race over the same reader,
  and callers that only check `err` leak the body. The contract (response
  returned alongside error) is unusual and undocumented.
  Fix: document the dual-return contract explicitly, or buffer the body into
  `RequestError` and close `res.Body` before returning.

- [correctness] `whatsapp/message.go:173-175` — `SendMessageResponse.Ok()`
  returns true if `FirstId()` is non-empty *or* `Success` is true. The Meta
  messages endpoint does not return a `success` field; an error response
  `{"error": {...}}` decodes into a zero-valued struct, `Ok()` is false — ok —
  but a partial/odd payload with `success:true` and no message id would be
  reported as sent. The `Success` field appears speculative.
  Fix: drop the unused `Success` field or document which endpoint sets it.

- [correctness] `whatsapp/media.go:24-52` — `GenerateMediaID` decodes the
  response into `Media` but never checks whether `media.Id` is empty. A Meta
  error body that is not `>= 400` (rare, but Graph API sometimes 200s with an
  error envelope) yields `("", nil)` — a silent failure.
  Fix: return an explicit error when `media.Id == ""`.

- [correctness] `whatsapp/message_template.go:104-117` —
  `DeleteMessageTemplate` reads nothing from the body and does not verify the
  API actually deleted the template; any 2xx is treated as success.
  Acceptable, but the function gives false confidence. Note `CreateMessage`
  `Template` drains the body but also ignores its content.
  Fix: decode and check the `success`/`error` envelope, or document the
  best-effort semantics.

- [correctness] `whatsapp/schedule_message_templates.go:25-26` —
  `time.LoadLocation(opts.Timezone)` error is discarded with `_`. On an
  invalid timezone string `loc` is `nil`, and `opts.StartTime.In(nil)` panics
  at runtime.
  Fix: handle the error — fall back to `time.UTC` (or return an error from
  `SchedulingTemplate`).

- [naming] `whatsapp/message_template.go:145` — exported constructor `Aproved()`
  is misspelled (should be `Approved`). It is part of the public API, so the
  typo is now load-bearing; renaming is a breaking change.
  Fix: add a correctly spelled `Approved()` and deprecate `Aproved()`.

- [testing] `whatsapp/upload_template_file.go` (93.3% / 94.7%) and
  `uuid/uuid.go:NewV7` (92.9%) — the `rand.Read` failure branch in `NewV7` and
  the `mediaFile.Open()` / `io.ReadAll` error branches in `UploadTemplateFile`
  are uncovered. These are the exact failure paths most likely to regress.
  Fix: inject a failing reader / fake `multipart.FileHeader` to cover them.

### Low

- [clarity] `whatsapp/template_analytics.go:424` — `return err == nil` is
  reached only when `err == nil`, so it always returns `true`. Dead branch.

- [clarity] `xhttp/xjson/json.go:19-22,41-43` — `Bytes()` and the package
  `Bytes()` swallow `json.Marshal` errors (`b, _ :=`). Documented as
  intentional, but a marshal failure silently yields `nil`/`""`; for a public
  lib an error-returning variant would be safer.

- [naming] `whatsapp/message.go:11-14` — `Header` is aliased to `xjson.JSON`
  here, while `xhttp` also exports a `Header` type (alias of `Data`). Two
  unrelated `Header` types across sibling packages is confusing.

- [api-stability] `whatsapp/message_template.go:132` — `status` is an
  *unexported* type returned by *exported* functions (`Pending()`, `Rejected()`,
  `Aproved()`) and embedded in the exported `TemplateStatus.Status` field.
  Callers cannot name the type. Export it or use a plain `string`.

- [dry] `whatsapp/*.go` — the pattern "build `xhttp.Options` with
  `Headers("application/json")`, POST, `defer Close`, `xjson.Decode`" is
  repeated in ~10 files. A small `postJSON[T]` helper would remove the
  boilerplate and centralize the body-close discipline (which is currently
  inconsistent — see the High findings).

- [clarity] `whatsapp/message_type.go:29-33` — `NewMessageType` linearly scans
  the `ContentTypes` map with a loop that does an equality check; this is just
  a map lookup written the slow way. Use `contentTypes[contentType]`.

- [consistency] `whatsapp/calling.go:33` vs `:33` — `NewCall` variadic param
  is named `logger ...func(data string)` but the field/closure type elsewhere
  is `func(message string)`; harmless but inconsistent naming.

- [version-control] `go.mod` declares `go 1.26.2` — a `go <major>.<minor>.<patch>`
  directive pins a very new toolchain; ensure CI and all contributors have it,
  otherwise builds fail with a toolchain-download surprise.

- [docs] Several files mix English doc comments with stray double spaces and a
  trailing-space issue (`typing_indicator.go`, `webhook_subscribe.go`); cosmetic
  only, `gofmt` would catch trailing whitespace inside comments only partially.

## Strengths
- Excellent automated-test coverage (99.6–100% across packages) with a
  purpose-built `xhttp.MockClient` — tests are network-free and deterministic.
- Clean layering: `xhttp` (transport) → `meta` (graph client) → `whatsapp`
  (domain) with the HTTP client behind the `HTTPClient` interface, so the
  domain code is testable without real sockets.
- `splitContactName` (the historical "invalid contact name" bug) is now
  resiliently handled: space → CamelCase → snake_case → single-word fallback.
- `Retry` correctly honors `context.Context` cancellation and reports the
  attempt count plus the last error.
- `uuid.NewV7` is a correct, allocation-light RFC 9562 implementation.
- `.gitignore` properly excludes `.env`, coverage artifacts and binaries; no
  hardcoded secrets found in source or tests.
- `MultipartWriter` consistently checks and propagates every `Write*`/`Close`
  error in `fileWriter`.

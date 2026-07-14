# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

> This service lives in the `ecommerce` multi-repo workspace. Workspace-wide rules
> (release-then-bump, `go.work`, hexagonal + `fx` conventions, tenancy, the gen/secrets
> edit guards) are in the **root `CLAUDE.md`** — read that first. This file only covers
> what is specific to `ecommerce-image-service`.

## What this service does

Image upload, storage, and delivery for the platform. It is **not** a CQRS query/read model —
it owns its own MongoDB `image` collection as the source of truth and coordinates three external
systems: **S3-compatible object storage** (MinIO locally, Cloudflare R2 in prod), **imgproxy**
(on-the-fly resizing/format conversion via signed URLs), and **Kafka** (consumes catalog product
events, emits image events via the transactional outbox).

## Commands

Standard workspace targets (`make run|build|test|lint|fmt|generate-mocks|check-all`) apply — see
root CLAUDE.md. Notable extras in this repo's `Makefile`:

```bash
make test-integration    # -tags=integration (real deps via testcontainers)
make bench               # go test -bench=. -benchmem
make profile-cpu         # CPU pprof over benchmarks, opens http://:8081
make docker-run          # run the image with --env-file .env on :8080
```
Run a single test: `go test ./internal/application/image/ -run TestConfirmUpload -v`

There is **no `-api` protobuf source in this repo**. Contracts come from the separate
`ecommerce-image-service-api` (RPCs it serves) and `ecommerce-catalog-service-api` (events it
consumes) modules. To change an RPC or event schema, edit `.proto` in the relevant `-api` repo and
`make generate` **there**, not here.

## Image lifecycle (the core domain flow)

The `Image` aggregate (`internal/application/image/image.go`) moves through statuses
`uploaded → ready → deleted` and owner types `draft | product | user`. The end-to-end flow:

1. **CreatePresign** — client asks for an upload slot. Handler generates the S3 key (drafts go to
   `drafts/<ownerId>/…`; product/user go to `<tenantSlug>/<prefix>/<ownerId>/…`), returns a
   presigned PUT URL **with Content-Type and Content-Length baked into the signature**, plus a
   signed **upload JWT** carrying the key/owner/size metadata.
2. **ConfirmUpload** — client calls back with the upload token. Handler validates the token,
   `HeadObject`s S3 to verify the file exists and its size matches the token (mismatch ⇒ delete +
   reject), then inserts the `Image` row. Drafts self-expire via a Mongo TTL index (see migration
   `000004`, 1800s on `ownerType=draft`).
3. **PromoteImages** — moves draft images to a real owner (product/user). This is a hand-rolled
   **saga**, not a single transaction (`promote_images.go`): Phase 1 copies S3 objects (originals
   kept); Phase 2 is a Mongo transaction that soft-deletes the owner's old images, updates promoted
   rows, and writes outbox messages; Phase 3 (post-commit, best-effort) deletes source + replaced
   files and flushes the outbox. On Phase 2 failure it **compensates by deleting the copied files**.
   Promotion is **idempotent** — already-promoted images to the same owner are skipped.
4. **GetDeliveryUrl** — builds a signed imgproxy URL from the image key + transform options; refuses
   deleted images.

## Kafka is the second entry point

Alongside the Connect-RPC handler, `internal/infrastructure/inbound/kafka/product_handler.go`
consumes `catalog.product.events`:
- `ProductUpdatedEvent` with an `ImageId` ⇒ **promote** that image to the product.
- `ProductUpdatedEvent` with no `ImageId`, or `ProductDeletedEvent` ⇒ **cleanup** (soft-delete rows
  + delete S3 objects) for that product owner.

So the same promote/cleanup use cases are driven both synchronously (RPC) and asynchronously
(events). Keep both paths in mind when changing those handlers.

## Two kinds of token, don't confuse them

- **Upload token** (`token_service.go`, impl in `outbound/security/jwt_token_service.go`) — an
  HS256 JWT this service *mints and verifies itself*, signed with `upload-token.jwt-secret`. It is
  the trust bridge between CreatePresign and ConfirmUpload; it is **not** end-user auth.
- **JWKS auth** — standard platform JWT validated against Logto's JWKS (`security.jwks`), wired via
  commons. This is caller authentication, same as every other service.

## Layout specifics

- `internal/application/image/` — domain + all use cases; port interfaces are defined here
  (`repository.go`, `object_storage.go`, `imgproxy_signer.go`, `token_service.go`, `Presigner` in
  `create_presign.go`). `mime_types.go` is the allowlist gate for accepted content types.
- `internal/infrastructure/outbound/` — the adapters implementing those ports: `s3/` (object
  storage + presigner + `tenant_cleaner.go`), `imgproxy/` (URL signer), `mongo/` (repository +
  entity/mapper), `security/` (upload-token JWT), `kafka/` (outbox event producer).
- `db/migrations/*.json` — **MongoDB index migrations as JSON command documents** (golang-migrate
  mongodb driver), run at startup because `main.go` wires `tenant.NewModule(tenant.WithMigrations())`.
  Add index changes as a new numbered up/down pair; do not mutate existing ones.
- `s3.NewImageTenantCleaner` is registered into the `tenant_cleaners` fx group so tenant deletion
  wipes that tenant's S3 prefix — remember it when changing key layout.

## Config knobs worth knowing

`configs/config.standalone.yaml` (overridable by env). Beyond the usual mongo/kafka/security:
- `application.max-upload-bytes` — enforced at both presign and confirm.
- `application.small-width` / `large-width` / `quality` — the two variant sizes emitted in
  promotion events and the imgproxy quality.
- `s3.presign-ttl` — also becomes the upload-token TTL.
- `imgproxy.key-hex` / `salt-hex` — HMAC keys for signed delivery URLs (dev values are placeholders).

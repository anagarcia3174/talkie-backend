# Talkie Backend

A Go API for **timestamp-specific commenting on movies and TV shows**. Users
search or browse trending media, land on a single media page, and post or
read comments anchored to a specific point in the runtime — "at 12:34, this
happens" — instead of a generic review or discussion thread.

---

## Core Feature

- Search or browse trending movies/TV via [TMDB](https://www.themoviedb.org/).
- A single media page per movie or show. TV shows use a season/episode
  number picker on that same page — there is no separate per-episode page.
- Post a comment anchored to a timestamp (in seconds) on a movie or a
  specific episode.
- Read comments for a piece of media, sorted by timestamp.
- Edit or delete your own comments.

## Explicitly Not in Scope (for now)

Reviews, ratings, watch-lists/watched status, a social graph (following,
likes), and OAuth. These were cut for scope discipline, not
because they're bad ideas — they're what caused the previous version to
stall.

## Moderation

- **Content filter:** comment bodies and usernames are checked against a
  small slur / self-harm list in `internal/moderation` (whole-word, with
  leetspeak and spacing normalization). Rejections return `422`.
- **Bans:** a banned user gets `403 account suspended` on create/edit
  comment, report, and `PATCH /me`, and their comments are hidden from
  comment lists. They can still delete their comments and their account.
  There's no admin UI yet — ban/unban from the Supabase SQL editor (see below).
- **Auto-hide:** a comment with 2+ reports that aren't `dismissed` is hidden
  from comment lists. To restore it, dismiss its reports (see below).
- **Blocking:** users can block other users. Comment list endpoints accept an
  optional JWT; when present, comments from users the viewer blocked are
  left out. A missing or invalid token is treated as signed out, never 401.

```sql
-- ban
UPDATE public.profiles SET banned_at = now() WHERE username = '<username>';
-- unban
UPDATE public.profiles SET banned_at = NULL WHERE username = '<username>';
-- restore an auto-hidden comment
UPDATE public.reports SET report_status = 'dismissed' WHERE comment_id = '<comment_id>';
```

---

## Tech Stack

| Concern | Choice |
|---|---|
| Language | Go |
| Routing | stdlib `net/http` + [`chi`](https://github.com/go-chi/chi) |
| DB access | [`sqlc`](https://sqlc.dev/) (typed Go generated from raw SQL) + `pgx/v5` |
| Migrations | [`goose`](https://github.com/pressly/goose) |
| Auth | email/password (`bcrypt`) + JWT (`golang-jwt/jwt/v5`) |
| Rate limiting | [`httprate`](https://github.com/go-chi/httprate) |
| External data | [TMDB API](https://developer.themoviedb.org/) (Bearer-token "API Read Access Token" auth) |
| Logging | `log/slog`, JSON handler |

No ORM, no framework beyond routing — SQL is written by hand and sqlc
generates the Go types and query methods around it.

---

## API Overview

### Auth
| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/signup` | — | Create an account |
| POST | `/login` | — | Exchange credentials for a JWT |
| GET | `/me` | required | Get the current user |
| GET | `/me/blocks` | required | List users the current user has blocked |
| POST | `/users/{user_id}/block` | required | Block a user (204; 400 self, 404 unknown user; repeat blocks are no-ops) |
| DELETE | `/users/{user_id}/block` | required | Unblock a user (204) |

### Media (proxies TMDB — read-only, no API key exposed to clients)
| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/search?query=` | — | Search movies and TV shows |
| GET | `/trending` | — | Daily trending movies and TV shows |
| GET | `/movies/{tmdb_id}` | — | Movie details |
| GET | `/tv/{tmdb_id}` | — | Show details + season picker data |
| GET | `/tv/{tmdb_id}/season/{season}/episode/{episode}` | — | Episode runtime + air date |

### Comments
| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/movies/{tmdb_id}/comments` | optional | List comments on a movie, by timestamp |
| GET | `/tv/{tmdb_id}/season/{season}/episode/{episode}/comments` | optional | List comments on an episode, by timestamp |
| POST | `/comments` | required | Post a comment |
| PATCH | `/comments/{comment_id}` | required, owner only | Edit a comment's body/timestamp |
| DELETE | `/comments/{comment_id}` | required, owner only | Delete a comment |
| POST | `/comments/{comment_id}/reports` | required | Report a comment (`reason`: spam, harassment, inappropriate, spoiler, other; optional `detail`, max 500 chars). 201 on success, 404 unknown comment, 403 own comment, 409 already reported |

**Reading is intentionally open.** Only posting, editing, and deleting a
comment require a JWT. This is a deliberate product decision: Talkie's
growth depends on shareable links and letting someone see a comment thread
before they're asked to sign up. Abuse protection on the open routes is
handled via per-IP rate limiting rather than an auth wall.

---

## Project Structure

```
talkie-backend/
├── cmd/api/                  # entrypoint, wiring, graceful shutdown
├── internal/
│   ├── auth/                 # signup/login/JWT, password hashing
│   ├── comments/             # comment CRUD + ownership rules
│   ├── config/                # env loading
│   ├── db/
│   │   ├── generated/        # sqlc output — do not hand-edit
│   │   └── queries/          # hand-written SQL, sqlc's input
│   ├── httpapi/              # HTTP handlers, router, middleware
│   ├── logging/
│   └── tmdb/                 # TMDB API client + normalized types
├── migrations/                # goose migrations
└── docker-compose.yml         # local Postgres + pgAdmin
```

Each layer owns its own shape of the data: TMDB's raw JSON, the DB's
generated row types, and the domain types used by services and handlers
are all kept deliberately separate rather than reused across layers.

---

## Design Principles

- **Read is free, write requires auth.** Browsing (search, trending, media
  details, comment lists) never requires a JWT — only actions that change
  data do.
- **Rate limiting over auth walls for abuse protection** on open routes —
  keeps discovery low-friction while still bounding scraping/abuse.
- **Comment edits are narrow.** Only a comment's `body` and
  `timestamp_seconds` can be changed. Changing what a comment is *about*
  means deleting and reposting, not editing.
- **No premature validation coupling.** Comment timestamps are checked for
  `>= 0` at write time; validating them against the media's actual runtime
  is left to the frontend, so posting a comment never depends on TMDB being
  up.
- **Errors are collapsed where enumeration matters, distinct where it
  doesn't.** Auth failures (bad email, bad password, invalid token) all
  return the same generic "invalid credentials" response, so an attacker
  can't use error differences to enumerate valid accounts. Comment errors
  don't have that concern, so "not found" (404) and "not yours" (403) are
  reported distinctly.
---

## Deploy (Fly.io)

Production runs on Fly.io as app `talkie-api` (`fly.toml`, `Dockerfile`),
served at `https://api.gettalkie.app`. The database is the Supabase project
in `DATABASE_URL`.

- **Config:** non-secret settings (`PORT`, `LOG_LEVEL`, `ALLOWED_ORIGIN`,
  `CLIENT_IP_HEADER`) are in `fly.toml` `[env]`. Secrets (`DATABASE_URL`,
  `TMDB_API_KEY`, `SUPABASE_URL`, `SUPABASE_SECRET_KEY`, `APPLE_*`) are set
  with `fly secrets set` / `fly secrets import`. `.env` is excluded from the
  image by `.dockerignore`.
- **Client IP:** `CLIENT_IP_HEADER=Fly-Client-IP` makes rate limiting use the
  real client IP instead of Fly's proxy. Leave it unset locally. Only set it
  where every request passes through a proxy that sets that header.
- **Deploy:** `fly deploy` (one machine; `--ha=false` on the first deploy).
- **Migrations:** run goose against `DATABASE_URL` before deploying code that
  needs them: `goose -dir migrations postgres "$DATABASE_URL" up`.
- **Logs / status:** `fly logs`, `fly status`. Health check: `GET /health`.
- **Domain:** `fly certs add api.gettalkie.app` + a Cloudflare CNAME
  `api → talkie-api.fly.dev` set to **DNS only** (grey cloud).

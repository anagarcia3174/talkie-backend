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
likes), user blocking, and OAuth. These were cut for scope discipline, not
because they're bad ideas — they're what caused the previous version to
stall.

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
| GET | `/movies/{tmdb_id}/comments` | — | List comments on a movie, by timestamp |
| GET | `/tv/{tmdb_id}/season/{season}/episode/{episode}/comments` | — | List comments on an episode, by timestamp |
| POST | `/comments` | required | Post a comment |
| PATCH | `/comments/{comment_id}` | required, owner only | Edit a comment's body/timestamp |
| DELETE | `/comments/{comment_id}` | required, owner only | Delete a comment |

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
# Game Price Tracker

A Steam game price tracker. Search for a game, add it to your personal list, set a **target price**, and get notified when the game reaches it.

[Leer en español](README.es.md)

## Features

- Sign up and log in with JWT authentication.
- Game search powered by the Steam store.
- Personal list with an editable target price per game.
- A **background worker** that periodically checks prices on Steam and creates a notification when a price drops to the target.
- Notifications panel with an unread counter.

## Stack

| Layer | Technology |
|---|---|
| Backend | Go, [chi](https://github.com/go-chi/chi) |
| Database | PostgreSQL 16 (pgx / pgxpool) |
| Auth | JWT |
| Frontend | Plain HTML + JavaScript (`web/index.html`), served by the Go server |
| Local infra | Docker Compose |
| Deployment | Dockerfile (e.g. Railway) |

## How it works

```
Browser ──► API (chi + JWT) ──► PostgreSQL
                 ▲
Worker (goroutine) ──► Steam Store API
```

Each row in `games` is a *user ↔ game* relation with that user's target price. On every cycle the worker:

1. Takes the **unique** `steam_app_id` values, so Steam is queried once per game no matter how many users track it.
2. Compares the current Steam price with the stored one. If it **hasn't changed**, it does nothing, which prevents repeated notifications on every cycle.
3. If it changed and the new price is lower than or equal to a user's target, it creates a notification for that user.
4. Stores the new price.

## Running locally

Requirements: Go and Docker.

```bash
# 1. Environment variables
cp .env.example .env        # then fill in the values

# 2. Database
docker compose up -d db
for f in db/migrations/*.sql; do
  docker compose exec -T db psql -U postgres -d content_tracker < "$f"
done

# 3. API + frontend
go run ./cmd/api
```

Open `http://localhost:8080`. The Go server serves both the API and the `web/` folder.

With `ENV=dev` you can also serve the frontend separately (`cd web && python3 -m http.server 5500 --bind 127.0.0.1`) and open `http://localhost:5500`; CORS is enabled for that origin in dev only.

To run in production mode locally:

```bash
ENV=production go run ./cmd/api
```

### Environment variables

| Variable | Description |
|---|---|
| `PORT` | API port (e.g. `8080`). On most platforms it is injected for you |
| `DATABASE_URL` | PostgreSQL connection string |
| `JWT_SECRET_KEY` | Secret used to sign tokens. Generate one with `openssl rand -base64 32` |
| `JWT_EXPIRATION_HOURS` | Token lifetime in hours |
| `ENV` | `dev` enables a 10-second worker interval, CORS for `localhost:5500` and the debug endpoint. Anything else runs in production mode (1-hour interval) |

The app refuses to start if `DATABASE_URL` or `JWT_SECRET_KEY` is missing. A `.env` file is optional: if it doesn't exist, plain environment variables are used.

## API

Endpoints marked with 🔒 require `Authorization: Bearer <token>`.

| Method | Route | Description |
|---|---|---|
| `POST` | `/users` | Create an account |
| `POST` | `/users/login` | Log in, returns the token |
| `GET` | `/games/search?term=` 🔒 | Search games on Steam |
| `POST` | `/games` 🔒 | Add a game. Body: `{"steam_app_id": 367520, "target_price": 2.5}` |
| `GET` | `/games/user-games` 🔒 | List my games |
| `PATCH` | `/games/update` 🔒 | Change the target. Body: `{"steam_app_id": 367520, "target_price": 3}` |
| `DELETE` | `/games/{steamAppID}` 🔒 | Remove a game from my list |
| `GET` | `/notifications` 🔒 | My notifications, newest first |
| `PATCH` | `/notifications/read` 🔒 | Mark my notifications as read |

Notable responses: adding a game that is already on the list returns `409 Conflict`, and adding one whose current price is already at or below the target returns `422 Unprocessable Entity`.

With `ENV=dev` there is also `POST /debug/check-prices`, which triggers a price check manually.

## Deployment

The repository includes a `Dockerfile`, so it can be deployed on any platform that builds from one (Railway, Render, ...):

1. Create a PostgreSQL database and run the migrations in `db/migrations/` against it.
2. Set `DATABASE_URL`, a **new** `JWT_SECRET_KEY`, `JWT_EXPIRATION_HOURS` and `ENV=production` in the platform's environment settings. Don't define `PORT` if the platform injects it.
3. Make sure the service stays always on: the price worker runs inside the same process as the API, so if the service sleeps, price checks stop.

## Design decisions

- **A single connection pool** (`pgxpool`) shared by handlers and the worker, instead of one shared connection.
- **Prices converted in one place:** Steam returns prices in cents, so they are converted to currency units when the response is read and the rest of the code always uses the same unit.
- **Per-row authorization:** update and delete queries also filter by `user_id`, so a user can't touch someone else's games.
- **One game per user:** `UNIQUE (user_id, steam_app_id)` prevents duplicates, and the API reports a clear conflict.
- **Worker in the same process:** a goroutine driven by `time.Ticker`. That's enough at this scale; it can be moved to a separate process if it grows.

## Roadmap/New features I want to add

- Price history (`price_history`) and a per-game chart.
- Email notifications.
- Notify on add when the game is already below the target.
- "Any price drop" alerts with no target.
- Normalized game catalog (separate game data from the user relation).
- Personalized recommendations.

## AI disclosure

No AI agents were used to build this project. Claude (an LLM) was used for debugging, explaining concepts, suggesting fixes and improvements, and generating the frontend.

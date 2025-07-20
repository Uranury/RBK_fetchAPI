# RBK fetch\_API

A lightweight Go monolith app that integrates with the [Steam Web API](https://developer.valvesoftware.com/wiki/Steam_Web_API) to fetch and expose structured data about users, games, and achievements.

---

## 🚀 Features

* 🔗 Resolve vanity URLs to Steam IDs
* 🎮 Fetch owned games for a Steam user
* 👤 Retrieve user profile summary
* 🏆 Get detailed game achievement data
* 📊 Multi-endpoint aggregation for achievement stats
* 📘 Swagger/OpenAPI documentation
* ⚠️ Graceful error handling with structured API responses

---

## 🛠 Tech Stack

* **Go** (Gin Framework) — HTTP routing and middleware
* **Redis** — Caching layer for performance
* **PostgreSQL** — Database for storing request history
* **sqlx** — Struct-mapped DB access
* **golang-migrate** — Migration tool
* **Docker** — Containerized development
* **Swagger** — Auto-generated API documentation
* **Steam Web API** — Data source
* **Postman** — API testing tool

---

## 📁 Folder Structure

```
internal/
  handlers/       # HTTP endpoints
  services/       # Business logic and Steam API
  apperrors/      # Centralized error handling
  models/         # API response data structures
  repositories/   # Database access layer
  server/         # App bootstrap logic
  db/
    migrations/   # DB schema migrations
```

---

## ⚙️ Environment Variables

Add the following variables to your `.env` file:

```env
STEAM_API_KEY=
LISTEN_ADDR=:8080
REDIS_ADDR=
POSTGRES_DSN=
POSTGRES_USER=
POSTGRES_PASSWORD=
POSTGRES_DB=
```

---

## 🧪 Running Tests

From the `services` directory:

```bash
go test -v
```

---

## 🖥 Run Locally

```bash
git clone https://github.com/Uranury/RBK_fetchAPI
cd RBK_fetchAPI
make build
```

This starts 3 containers: `db`, `myapp`, and `redis`.

---

## 📘 API Reference

### 🔎 `/steam_id` — Resolve Vanity URL

```http
GET /steam_id?vanity=eldenringmaster
```

#### Parameters

| Name   | Type   | Required | Description                   |
| ------ | ------ | -------- | ----------------------------- |
| vanity | string | Yes      | Custom Steam profile URL name |

#### Success Response

```json
{
  "steamID": "76561199054042527"
}
```

**Usage:**
Use this before calling `/achievements`, `/games`, or `/summary` that require `steamID64`.

---

### 🧑 `/summary` — Steam Profile Summary

```http
GET /summary?steam_id=76561198377031178
```

#### Parameters

| Name      | Type   | Required | Description               |
| --------- | ------ | -------- | ------------------------- |
| steam\_id | string | Yes      | 64-bit Steam ID of player |

#### Success Response

Returns public profile info:

```json
{
  "response": {
    "players": [
      {
        "steamid": "76561198377031178",
        "personaname": "인턴십",
        ...
      }
    ]
  }
}
```

---

### 🎮 `/games` — Owned Games

```http
GET /games?steam_id=76561198377031178
```

#### Parameters

| Name      | Type   | Required | Description               |
| --------- | ------ | -------- | ------------------------- |
| steam\_id | string | Yes      | 64-bit Steam ID of player |

#### Success Response

```json
{
  "response": {
    "game_count": 36,
    "games": [
      {
        "appid": 105600,
        "name": "Terraria",
        "playtime_forever": 6682,
        "has_community_visible_stats": true
      },
      ...
    ]
  }
}
```

#### Icon URL Format

```
https://media.steampowered.com/steamcommunity/public/images/apps/{appid}/{img_icon_url}.jpg
```

---

### 🏆 `/achievements` — Game Achievements for a User

```http
GET /achievements?appID=1245620&steamID=76561198377031178
```

#### Parameters

| Name    | Type   | Required | Description              |
| ------- | ------ | -------- | ------------------------ |
| steamID | string | Yes      | 64-bit Steam ID          |
| appID   | string | Yes      | Steam App ID of the game |

#### Success Response

```json
{
  "steamID": "76561198377031178",
  "gameName": "ELDEN RING",
  "achievements": [
    {
      "name": "ACH00",
      "displayName": "Elden Ring",
      "achieved": true,
      "rarity": 10.1
    },
    ...
  ]
}
```

---

## 📦 Example Use Cases

* Build user dashboards with achievements
* Show progress per game
* Sort by rarity or playtime
* Display profile widgets

---

## 🔗 Useful Links

[![GitHub](https://img.shields.io/badge/github-181717?style=for-the-badge\&logo=github\&logoColor=white)](https://github.com/Uranury)
[![LinkedIn](https://img.shields.io/badge/linkedin-0A66C2?style=for-the-badge\&logo=linkedin\&logoColor=white)](https://www.linkedin.com/in/alibi-ulanuly-37700330b/)

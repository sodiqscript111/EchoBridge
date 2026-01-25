<p align="center">
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go Version" />
  <img src="https://img.shields.io/badge/Svelte-5-FF3E00?style=for-the-badge&logo=svelte&logoColor=white" alt="Svelte" />
  <img src="https://img.shields.io/badge/PostgreSQL-16-4169E1?style=for-the-badge&logo=postgresql&logoColor=white" alt="PostgreSQL" />
  <img src="https://img.shields.io/badge/Temporal-Latest-7C3AED?style=for-the-badge&logo=temporal&logoColor=white" alt="Temporal" />
</p>

<h1 align="center">EchoBridge</h1>

<p align="center">
  <strong>Seamlessly transfer your music playlists across streaming platforms</strong>
</p>

<p align="center">
  Spotify | YouTube Music
</p>

---

## Overview

EchoBridge is a playlist synchronization service that enables users to transfer their music libraries between Spotify and YouTube Music. The system handles the complexity of cross-platform track matching, rate limiting, and fault-tolerant operations through a distributed workflow architecture.

---

## System Design

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Frontend (SvelteKit)                     │
│              Responsive UI • OAuth Flow • Real-time Updates      │
└─────────────────────────────────────────────────────────────────┘
                                   │
                                   ▼
┌─────────────────────────────────────────────────────────────────┐
│                        Go Backend (Gin)                          │
│         REST API • Auth Middleware • Request Handling            │
└─────────────────────────────────────────────────────────────────┘
                    │                        │
                    ▼                        ▼
┌──────────────────────────────┐  ┌────────────────────────────────┐
│     Temporal Workflows       │  │        Worker Pool             │
│  • PlaylistSyncWorkflow      │  │  Concurrent task processing    │
│  • ImportPlaylistWorkflow    │  │  Background categorization     │
│  • Rate limit handling       │  │                                │
└──────────────────────────────┘  └────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────────────────────┐
│                  Platform Services                               │
│            Spotify API  •  YouTube Data API                      │
└─────────────────────────────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────────────────────┐
│                    PostgreSQL Database                           │
│         Users • Playlists • Tracks • OAuth Tokens                │
└─────────────────────────────────────────────────────────────────┘
```

### Core Components

#### 1. Authentication Layer
The system uses OAuth 2.0 for platform authentication. When a user connects a streaming service:
1. User initiates OAuth flow via `/api/link/{platform}`
2. After authorization, the callback handler receives tokens
3. Tokens are encrypted and stored in PostgreSQL
4. JWT tokens are issued for API authentication

#### 2. Playlist Import Flow
```
User Request → Backend validates token → Platform Service fetches playlist
     │
     ▼
Temporal ImportPlaylistWorkflow starts
     │
     ├── FetchUserActivity (retrieves user from DB)
     ├── Paginated track fetching from source platform
     ├── Track normalization (title, artist, album extraction)
     └── Batch insert to PostgreSQL
```

The import is handled as a Temporal workflow, which provides:
- **Durability**: If the server crashes mid-import, the workflow resumes from the last checkpoint
- **Retry logic**: Failed API calls are automatically retried with exponential backoff
- **Visibility**: Progress can be queried at any point

#### 3. Playlist Sync Flow
The sync operation transfers tracks from an imported playlist to target platforms:

```
PlaylistSyncWorkflow
     │
     ├── FetchPlaylistActivity (load playlist + tracks from DB)
     │
     └── For each target platform:
          │
          ├── CreatePlaylistActivity (create empty playlist on target)
          │
          └── For each track:
               │
               ├── SearchTrackActivity (find matching track on target)
               │    └── Uses title + artist fuzzy matching
               │
               └── AddTrackActivity (add to target playlist)
                    └── Handles rate limits with automatic backoff
```

#### 4. Rate Limit Handling
Streaming platforms enforce strict rate limits. The system handles this through:

1. **Temporal Sleep**: When a rate limit is detected, the workflow sleeps for the reset duration
2. **Automatic Resume**: After the sleep, processing continues from exact point of interruption
3. **Exponential Backoff**: Retry policies use `2^n` second intervals (1s, 2s, 4s, 8s...)

```go
RetryPolicy: &temporal.RetryPolicy{
    InitialInterval:    time.Second,
    BackoffCoefficient: 2.0,
    MaximumInterval:    time.Minute,
    MaximumAttempts:    5,
}
```

#### 5. Worker Pool
For operations that don't require Temporal's durability guarantees, a lightweight worker pool handles concurrent tasks:
- **Pool Size**: Configurable (default 100 workers)
- **Active Workers**: 5 goroutines pull from the job queue
- **Use Cases**: Background categorization, non-critical updates

### Data Model

```
┌──────────────┐       ┌──────────────┐       ┌──────────────┐
│    Users     │       │   Playlists  │       │    Tracks    │
├──────────────┤       ├──────────────┤       ├──────────────┤
│ id           │──┐    │ id           │──┐    │ id           │
│ email        │  │    │ owner_id     │←─┘    │ playlist_id  │←─┐
│ username     │  │    │ title        │  │    │ title        │  │
│ spotify_token│  └───→│ description  │  │    │ artist       │  │
│ youtube_token│       │ platform     │  │    │ album        │  │
│              │       │ source_id    │  └───→│ spotify_id   │  │
│              │       │ is_public    │       │ youtube_id   │  │
└──────────────┘       └──────────────┘       │ duration_ms  │  │
                                              └──────────────┘  │
                                                                │
                                              (Foreign Key)─────┘
```

### Why Temporal?

Traditional approaches to playlist sync face several challenges:

| Problem | Traditional Solution | Temporal Solution |
|---------|---------------------|-------------------|
| Server crash mid-sync | Data loss, manual restart | Automatic resume from checkpoint |
| Rate limit exceeded | Fail entire operation | Sleep and resume transparently |
| Network timeout | Retry logic in application | Built-in retry policies |
| Progress tracking | Custom database state | Query workflow state anytime |
| Long-running operations | Background jobs + polling | Durable execution with callbacks |

---

## Tech Stack

| Layer | Technology |
|-------|------------|
| **Backend** | Go, Gin Framework |
| **Frontend** | SvelteKit, TypeScript |
| **Database** | PostgreSQL, GORM |
| **Workflows** | Temporal |
| **Auth** | OAuth 2.0, JWT |
| **APIs** | Spotify Web API, YouTube Data API |

---

## Getting Started

### Prerequisites

- Go 1.21+
- Node.js 18+
- PostgreSQL 16+
- Temporal (optional, for workflow features)

### Installation

1. Clone the repository
   ```bash
   git clone https://github.com/sodiqscript111/EchoBridge.git
   cd EchoBridge
   ```

2. Configure environment variables
   ```bash
   cp .env.example .env
   ```

3. Set required values in `.env`:
   ```env
   DATABASE_URL=postgres://user:password@localhost:5432/echobridge
   GOOGLE_CLIENT_ID=your_google_client_id
   GOOGLE_CLIENT_SECRET=your_google_client_secret
   SPOTIFY_CLIENT_ID=your_spotify_client_id
   SPOTIFY_SECRET=your_spotify_secret
   JWT_SECRET=your_secure_jwt_secret
   ```

4. Run the application
   ```bash
   go run main.go
   ```

5. Optional: Start Temporal for workflow features
   ```bash
   temporal server start-dev
   ```

---

## License

This project is licensed under the MIT License.

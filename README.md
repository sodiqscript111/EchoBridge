EchoBridge: Cross-Platform Music Playlist Migration
Project Description
EchoBridge is a powerful and efficient web service built with Go, designed to eliminate the frustration of fragmented music libraries across different streaming platforms. It currently provides seamless one-way migration of playlists from Spotify to YouTube Music, with ambitious plans for future expansion to include Apple Music and a full-fledged playlist discovery platform. This project demonstrates proficiency in building robust API integrations, secure authentication flows, and scalable backend services.

Problem Solved
In today's diverse music streaming landscape, users often find their favorite playlists scattered across multiple services. Manually recreating these playlists is a tedious and time-consuming task. EchoBridge automates this process, allowing users to effortlessly transfer their curated music collections between platforms, ensuring their "banger" playlists are available wherever they choose to listen.

Current Features
Spotify Integration:

Secure OAuth 2.0 authentication flow.

Fetches the authenticated user's private and public playlists, including track details.

YouTube Music Integration:

Secure OAuth 2.0 authentication flow with necessary scopes for write operations.

Ability to search for corresponding music videos on YouTube based on Spotify track metadata (title, artist).

Creates new private YouTube playlists (or identifies existing ones by name).

Adds individual music videos as playlist items to the target YouTube playlist.

Playlist Export Automation: Automates the end-to-end process of pulling a Spotify playlist and populating a YouTube playlist.

Pagination Handling: Efficiently processes large playlists by automatically fetching all pages of tracks from Spotify.

Detailed Server-Side Logging: Provides real-time feedback on the export process, including successful additions, skipped tracks (if no YouTube match is found), and failures.

Technical Stack & Architecture
Backend Framework: Go (net/http, log, strings) with Gin Gonic for a high-performance RESTful API.

Authentication: Implementation of OAuth 2.0 Authorization Code Flow using golang.org/x/oauth2 for Google (YouTube Data API) and github.com/zmb3/spotify/v2/auth for Spotify. This ensures secure user consent and token management.

API Clients: Utilizes official Go SDKs for seamless interaction with the Spotify Web API (github.com/zmb3/spotify/v2) and YouTube Data API v3 (google.golang.org/api/youtube/v3).

Data Handling: In-memory global storage for authenticated tokens and API clients (for demonstration purposes – see Future Enhancements for production-ready approach).

Core Logic: Handles HTTP requests, manages authentication redirects, performs cross-API calls for search and playlist modifications, and implements iterative processing for playlist content.

Getting Started
To set up and run EchoBridge locally, follow these steps:

Prerequisites
Go: Ensure you have Go installed (version 1.16 or higher is recommended).

Spotify Developer Account:

Create an application on the Spotify Developer Dashboard.

Note down your Client ID and Client Secret.

Set the Redirect URI for your Spotify application to http://127.0.0.1:8000/callback/spotify.

Google Cloud Project:

Create a project in the Google Cloud Console.

Enable the YouTube Data API v3 for your project (APIs & Services > Enabled APIs & Services).

Create OAuth 2.0 Client ID credentials for a "Web application" (APIs & Services > Credentials).

Note down your Client ID and Client Secret.

Set the Authorized redirect URI for your OAuth client to http://127.0.0.1:8000/callback/youtube.

Crucially, configure your OAuth Consent Screen and ensure your Google account is added as a "Test user" to allow testing.

Configuration
Clone the Repository (if you haven't yet):

git clone https://github.com/sodiqscript111/EchoBridge.git
cd EchoBridge

Update main.go with Credentials:
Open main.go and replace the placeholder spotifyClientID, spotifyClientSecret, youtubeClientID, and youtubeClientSecret with your actual credentials from the respective developer dashboards.

// In func init()
// --- Spotify Configuration ---
spotifyClientID := "YOUR_SPOTIFY_CLIENT_ID"
spotifyClientSecret := "YOUR_SPOTIFY_CLIENT_SECRET"

// --- YouTube Music Configuration ---
youtubeClientID := "YOUR_YOUTUBE_CLIENT_ID"
youtubeClientSecret := "YOUR_YOUTUBE_CLIENT_SECRET"

Ensure Correct YouTube Scope:
Verify that youtube.YoutubeScope is included in the Scopes array for youtubeOAuthConfig in func init():

Scopes: []string{
    youtube.YoutubeReadonlyScope,
    youtube.YoutubeScope, // This is essential for creating/modifying playlists
},

Running the Application
Install Dependencies:

go mod tidy

Start the Server:

go run main.go

The server will start on http://127.0.0.1:8000.

Testing the Export Functionality
Authenticate Spotify:

Open your browser and go to: http://127.0.0.1:8000/login/spotify

Authorize the application. Copy the id of a Spotify playlist from the JSON response.

Authenticate YouTube (Crucial Re-authorization if Scopes Changed!):

Open your browser and go to: http://127.0.0.1:8000/login/youtube

Carefully review and grant the requested permissions, especially "Manage your YouTube videos and playlists." This step is vital for the export functionality to work.

Trigger Playlist Export:

In your browser, navigate to: http://127.0.0.1:8000/export/spotify/YOUR_SPOTIFY_PLAYLIST_ID/to/youtube

Replace YOUR_SPOTIFY_PLAYLIST_ID with the ID you copied from Spotify.

Verify:

Observe the logs in your terminal for real-time export progress.

Check your YouTube Music account (under "Library" -> "Playlists") for the newly created or updated playlist with the same name as your Spotify playlist.

Future Enhancements & Roadmap
EchoBridge is envisioned as a comprehensive music management solution. Planned features include:

Apple Music Integration: Adding full support for Apple Music as both a source and destination platform.

Bi-directional Sync: Expand functionality to allow seamless playlist transfers between any two supported platforms (e.g., YouTube to Spotify, Apple Music to YouTube).

Playlist Discovery Platform: Build a community-driven web user interface where users can share, discover, and export curated playlists created by others directly to their streaming service of choice. This involves:

Full user authentication system for EchoBridge accounts.

Database integration for persistent storage of user data, shared playlist metadata, and tokens.

Advanced search and filtering for discovering playlists by genre, mood, or creator.

User profiles to showcase their shared playlists.

Robust Frontend UI: Develop an intuitive and interactive web interface using modern frontend technologies (e.g., React, Vue, or simple HTML/JS served by Go) for a superior user experience.

Asynchronous Processing & Scalability: Implement Goroutines and proper rate limiting for parallel track searching and adding, enhancing performance for larger playlists and ensuring adherence to API quotas.

Advanced Track Matching: Explore more sophisticated algorithms or even machine learning techniques for higher accuracy in finding equivalent songs across diverse music catalogs.

Token Refresh & Management: Implement robust mechanisms for refreshing OAuth tokens automatically and securely storing them in a database.

Contribution
Ideas and contributions are welcome! Feel free to open issues or pull requests.

Contact

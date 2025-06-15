package main

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

var (
	spotifyRedirectURL = "http://127.0.0.1:8000/callback/spotify"
	spotifyAuth        *spotifyauth.Authenticator
	youtubeRedirectURL = "http://127.0.0.1:8000/callback/youtube"
	youtubeOAuthConfig *oauth2.Config
	state              = "random-state-string"
	spotifyUserToken   *oauth2.Token
	youtubeUserToken   *oauth2.Token
	spotifyAPIClient   *spotify.Client
	youtubeAPIClient   *youtube.Service
)

func init() {
	spotifyClientID := "6f55af09467d44319c84abf777042886"
	spotifyClientSecret := "879b791b755c4a9cbb97254bee4d3917"

	spotifyAuth = spotifyauth.New(
		spotifyauth.WithClientID(spotifyClientID),
		spotifyauth.WithClientSecret(spotifyClientSecret),
		spotifyauth.WithRedirectURL(spotifyRedirectURL),
		spotifyauth.WithScopes(
			spotifyauth.ScopeUserReadPrivate,
			spotifyauth.ScopePlaylistReadPrivate,
			spotifyauth.ScopePlaylistModifyPrivate,
			spotifyauth.ScopePlaylistModifyPublic,
		),
	)

	youtubeClientID := "1055231814715-mkfdshbi2l4rkmvn39l4f530o61o6c4u.apps.googleusercontent.com"
	youtubeClientSecret := "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"

	youtubeOAuthConfig = &oauth2.Config{
		ClientID:     youtubeClientID,
		ClientSecret: youtubeClientSecret,
		RedirectURL:  youtubeRedirectURL,
		Scopes: []string{
			youtube.YoutubeReadonlyScope,
			youtube.YoutubeScope,
		},
		Endpoint: google.Endpoint,
	}
}

func main() {
	r := gin.Default()

	r.GET("/login/spotify", spotifyLoginHandler)
	r.GET("/callback/spotify", spotifyCallbackHandler)
	r.GET("/spotify/playlist/:id/tracks", spotifyPlaylistTracksHandler)
	r.GET("/login/youtube", youtubeLoginHandler)
	r.GET("/callback/youtube", youtubeCallbackHandler)
	r.GET("/youtube/playlists", youtubePlaylistsHandler)
	r.GET("/export/spotify/:spotifyPlaylistID/to/youtube", exportSpotifyToYouTubeHandler)

	r.Run(":8000")
}

func spotifyLoginHandler(c *gin.Context) {
	url := spotifyAuth.AuthURL(state)
	c.Redirect(http.StatusFound, url)
}

func spotifyCallbackHandler(c *gin.Context) {
	r := c.Request

	token, err := spotifyAuth.Token(r.Context(), state, r)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Failed to obtain Spotify access token",
			"details": err.Error(),
		})
		return
	}

	spotifyUserToken = token
	spotifyAPIClient = spotify.New(spotifyAuth.Client(r.Context(), spotifyUserToken))

	user, err := spotifyAPIClient.CurrentUser(r.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve Spotify user profile",
			"details": err.Error(),
		})
		return
	}

	playlists, err := spotifyAPIClient.CurrentUsersPlaylists(r.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve Spotify playlists",
			"details": err.Error(),
		})
		return
	}

	var result []gin.H
	for _, p := range playlists.Playlists {
		result = append(result, gin.H{
			"platform": "spotify",
			"id":       p.ID.String(),
			"name":     p.Name,
			"tracks":   p.Tracks.Total,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Successfully authenticated with Spotify",
		"user":      user.ID,
		"token":     token.AccessToken,
		"playlists": result,
	})
}

func spotifyPlaylistTracksHandler(c *gin.Context) {
	playlistID := c.Param("id")
	authHeader := c.GetHeader("Authorization")

	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Missing or invalid Authorization header",
		})
		return
	}

	accessToken := strings.TrimPrefix(authHeader, "Bearer ")
	token := &oauth2.Token{AccessToken: accessToken}
	httpClient := spotifyAuth.Client(context.Background(), token)
	client := spotify.New(httpClient)

	tracksPage, err := client.GetPlaylistTracks(context.Background(), spotify.ID(playlistID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve Spotify playlist tracks",
			"details": err.Error(),
		})
		return
	}

	var tracks []gin.H
	for _, item := range tracksPage.Tracks {
		if item.Track.ID != "" {
			var artists []string
			for _, a := range item.Track.Artists {
				artists = append(artists, a.Name)
			}
			tracks = append(tracks, gin.H{
				"name":    item.Track.Name,
				"artists": artists,
				"album":   item.Track.Album.Name,
				"uri":     item.Track.URI,
				"id":      item.Track.ID.String(),
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"platform":    "spotify",
		"playlist_id": playlistID,
		"tracks":      tracks,
	})
}

func youtubeLoginHandler(c *gin.Context) {
	url := youtubeOAuthConfig.AuthCodeURL(state)
	c.Redirect(http.StatusFound, url)
}

func youtubeCallbackHandler(c *gin.Context) {
	r := c.Request

	if authError := r.URL.Query().Get("error"); authError != "" {
		errorDescription := r.URL.Query().Get("error_description")
		c.JSON(http.StatusForbidden, gin.H{
			"error":       "YouTube authorization failed",
			"details":     authError,
			"description": errorDescription,
			"hint":        "User likely denied access or there was an authorization issue.",
		})
		return
	}

	authCode := r.URL.Query().Get("code")
	if authCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Missing authorization code",
			"details": "The 'code' parameter was not found in the redirect URL.",
			"hint":    "Ensure the redirect URI in Google Cloud Console is correctly configured.",
		})
		return
	}

	token, err := youtubeOAuthConfig.Exchange(r.Context(), authCode)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Failed to obtain YouTube access token",
			"details": err.Error(),
		})
		return
	}

	youtubeUserToken = token
	httpClient := youtubeOAuthConfig.Client(context.Background(), youtubeUserToken)
	youtubeAPIClient, err = youtube.NewService(context.Background(), option.WithHTTPClient(httpClient))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to initialize YouTube service client",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully authenticated with YouTube",
		"token":   token.AccessToken,
	})
}

func youtubePlaylistsHandler(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")

	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Missing or invalid Authorization header",
		})
		return
	}

	accessToken := strings.TrimPrefix(authHeader, "Bearer ")
	token := &oauth2.Token{AccessToken: accessToken}
	httpClient := youtubeOAuthConfig.Client(context.Background(), token)

	service, err := youtube.NewService(context.Background(), option.WithHTTPClient(httpClient))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to initialize YouTube service",
			"details": err.Error(),
		})
		return
	}

	call := service.Playlists.List([]string{"id", "snippet", "contentDetails"}).Mine(true)
	response, err := call.Do()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve YouTube playlists",
			"details": err.Error(),
		})
		return
	}

	var result []gin.H
	if response != nil && response.Items != nil {
		for _, playlist := range response.Items {
			if playlist.Id != "" && playlist.Snippet != nil && playlist.Snippet.Title != "" {
				result = append(result, gin.H{
					"platform":    "youtube",
					"id":          playlist.Id,
					"title":       playlist.Snippet.Title,
					"description": playlist.Snippet.Description,
					"item_count":  playlist.ContentDetails.ItemCount,
				})
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Successfully retrieved YouTube playlists",
		"playlists": result,
	})
}

func searchYouTubeVideo(ctx context.Context, service *youtube.Service, title, artist string) (string, error) {
	searchQuery := title + " " + artist + " official music video"
	call := service.Search.List([]string{"id", "snippet"}).
		Q(searchQuery).
		MaxResults(1).
		Type("video")
	response, err := call.Do()
	if err != nil {
		return "", err
	}
	if len(response.Items) > 0 {
		return response.Items[0].Id.VideoId, nil
	}
	return "", nil
}

func createYouTubePlaylist(ctx context.Context, service *youtube.Service, playlistName, description string) (string, error) {
	playlist := &youtube.Playlist{
		Snippet: &youtube.PlaylistSnippet{
			Title:       playlistName,
			Description: description,
		},
		Status: &youtube.PlaylistStatus{
			PrivacyStatus: "private",
		},
	}
	call := service.Playlists.Insert([]string{"snippet", "status"}, playlist)
	response, err := call.Do()
	if err != nil {
		return "", err
	}
	return response.Id, nil
}

func addYouTubePlaylistItem(ctx context.Context, service *youtube.Service, playlistID, videoID string) error {
	playlistItem := &youtube.PlaylistItem{
		Snippet: &youtube.PlaylistItemSnippet{
			PlaylistId: playlistID,
			ResourceId: &youtube.ResourceId{
				Kind:    "youtube#video",
				VideoId: videoID,
			},
		},
	}
	call := service.PlaylistItems.Insert([]string{"snippet"}, playlistItem)
	_, err := call.Do()
	return err
}

func findYouTubePlaylistByName(ctx context.Context, service *youtube.Service, playlistName string) (string, error) {
	call := service.Playlists.List([]string{"id", "snippet"}).Mine(true)
	response, err := call.Do()
	if err != nil {
		return "", err
	}
	if response != nil && response.Items != nil {
		for _, playlist := range response.Items {
			if playlist.Snippet != nil && playlist.Snippet.Title == playlistName {
				return playlist.Id, nil
			}
		}
	}
	return "", nil
}

func exportSpotifyToYouTubeHandler(c *gin.Context) {
	spotifyPlaylistID := c.Param("spotifyPlaylistID")

	if spotifyAPIClient == nil || youtubeAPIClient == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required for both Spotify and YouTube. Please log in to both platforms.",
		})
		return
	}

	spotifyPlaylist, err := spotifyAPIClient.GetPlaylist(context.Background(), spotify.ID(spotifyPlaylistID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve Spotify playlist details",
			"details": err.Error(),
		})
		return
	}

	youtubePlaylistID, err := findYouTubePlaylistByName(context.Background(), youtubeAPIClient, spotifyPlaylist.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to check for existing YouTube playlist",
			"details": err.Error(),
		})
		return
	}

	if youtubePlaylistID == "" {
		youtubePlaylistID, err = createYouTubePlaylist(context.Background(), youtubeAPIClient, spotifyPlaylist.Name, spotifyPlaylist.Description)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to create YouTube playlist",
				"details": err.Error(),
			})
			return
		}
	}

	tracksPage, err := spotifyAPIClient.GetPlaylistTracks(context.Background(), spotify.ID(spotifyPlaylistID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve Spotify playlist tracks for export",
			"details": err.Error(),
		})
		return
	}

	var exportedTrackCount, failedTrackCount, skippedTrackCount int

	for page := 1; ; page++ {
		for _, item := range tracksPage.Tracks {
			if item.Track.ID == "" {
				skippedTrackCount++
				continue
			}

			trackName := item.Track.Name
			var artists []string
			for _, a := range item.Track.Artists {
				artists = append(artists, a.Name)
			}
			artistNames := strings.Join(artists, ", ")

			youtubeVideoID, err := searchYouTubeVideo(context.Background(), youtubeAPIClient, trackName, artistNames)
			if err != nil {
				failedTrackCount++
				continue
			}

			if youtubeVideoID != "" {
				err = addYouTubePlaylistItem(context.Background(), youtubeAPIClient, youtubePlaylistID, youtubeVideoID)
				if err != nil {
					failedTrackCount++
				} else {
					exportedTrackCount++
				}
			} else {
				skippedTrackCount++
			}
		}

		if tracksPage.Next == "" {
			break
		}
		tracksPage, err = spotifyAPIClient.GetPlaylistTracks(
			context.Background(),
			spotify.ID(spotifyPlaylistID),
			spotify.Offset(int(tracksPage.Offset)+len(tracksPage.Tracks)),
		)
		if err != nil {
			break
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":                    "Playlist export completed successfully",
		"spotify_playlist_name":      spotifyPlaylist.Name,
		"spotify_playlist_id":        spotifyPlaylistID,
		"youtube_playlist_id":        youtubePlaylistID,
		"total_spotify_tracks":       spotifyPlaylist.Tracks.Total,
		"tracks_exported_to_youtube": exportedTrackCount,
		"tracks_failed_to_export":    failedTrackCount,
		"tracks_skipped_no_youtube":  skippedTrackCount,
		"hint":                       "Check your YouTube account for the new or updated playlist.",
	})
}

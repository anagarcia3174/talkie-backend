package tmdb

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const baseURL = "https://api.themoviedb.org/3"

type Client struct {
	apiKey     string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) doGet(ctx context.Context, path string, query url.Values, dest any) error {
	if query == nil {
		query = url.Values{}
	}

	fullURL := baseURL + path
	if len(query) > 0 {
		fullURL += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return fmt.Errorf("tmdb: building request for %s: %w", path, err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("tmdb: request to %s failed: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("tmdb: %s returned status %d", path, resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(dest); err != nil {
		return fmt.Errorf("tmdb: decoding response from %s: %w", path, err)
	}

	return nil
}


type rawMultiSearchItem struct {
	ID           int    `json:"id"`
	MediaType    string `json:"media_type"`
	Title        string `json:"title"`
	Name         string `json:"name"`
	ReleaseDate  string `json:"release_date"`
	FirstAirDate string `json:"first_air_date"`
	PosterPath   string `json:"poster_path"`
	Overview     string `json:"overview"`
}

type rawSearchResponse struct {
	Results []rawMultiSearchItem `json:"results"`
}

type rawMovieDetails struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	ReleaseDate string `json:"release_date"`
	Runtime     int    `json:"runtime"`
	PosterPath  string `json:"poster_path"`
	Overview    string `json:"overview"`
}

type rawSeason struct {
	SeasonNumber int    `json:"season_number"`
	EpisodeCount int    `json:"episode_count"`
	AirDate      string `json:"air_date"`
}

type rawTVShowDetails struct {
	ID           int         `json:"id"`
	Name         string      `json:"name"`
	FirstAirDate string      `json:"first_air_date"`
	PosterPath   string      `json:"poster_path"`
	Overview     string      `json:"overview"`
	Seasons      []rawSeason `json:"seasons"`
}

type rawEpisodeDetails struct {
	Runtime int    `json:"runtime"`
	AirDate string `json:"air_date"`
}

type rawTrendingItem struct {
	ID           int    `json:"id"`
	MediaType    string `json:"media_type"`
	Title        string `json:"title"`
	Name         string `json:"name"`
	ReleaseDate  string `json:"release_date"`
	FirstAirDate string `json:"first_air_date"`
	PosterPath   string `json:"poster_path"`
	Overview     string `json:"overview"`
}

type rawTrendingResponse struct {
	Results []rawTrendingItem `json:"results"`
}

func yearFromDate(date string) int {
	if len(date) < 4 {
		return 0
	}
	var year int
	if _, err := fmt.Sscanf(date[:4], "%d", &year); err != nil {
		return 0
	}
	return year
}


func (c *Client) SearchMulti(ctx context.Context, query string) ([]SearchResult, error) {
	var raw rawSearchResponse
	q := url.Values{"query": {query}}
	if err := c.doGet(ctx, "/search/multi", q, &raw); err != nil {
		return nil, err
	}

	results := make([]SearchResult, 0, len(raw.Results))
	for _, item := range raw.Results {
		var mt MediaType
		switch item.MediaType {
		case "movie":
			mt = MediaTypeMovie
		case "tv":
			mt = MediaTypeTV
		default:
		
			continue
		}

		title := item.Title
		releaseDate := item.ReleaseDate
		if mt == MediaTypeTV {
			title = item.Name
			releaseDate = item.FirstAirDate
		}

		results = append(results, SearchResult{
			TMDBID:      item.ID,
			MediaType:   mt,
			Title:       title,
			ReleaseYear: yearFromDate(releaseDate),
			PosterPath:  item.PosterPath,
			Overview:    item.Overview,
		})
	}

	return results, nil
}

func (c *Client) GetMovieDetails(ctx context.Context, tmdbID int) (MovieDetails, error) {
	var raw rawMovieDetails
	path := fmt.Sprintf("/movie/%d", tmdbID)
	if err := c.doGet(ctx, path, nil, &raw); err != nil {
		return MovieDetails{}, err
	}

	return MovieDetails{
		TMDBID:         raw.ID,
		Title:          raw.Title,
		ReleaseYear:    yearFromDate(raw.ReleaseDate),
		RuntimeMinutes: raw.Runtime,
		PosterPath:     raw.PosterPath,
		Overview:       raw.Overview,
	}, nil
}

func (c *Client) GetTVShowDetails(ctx context.Context, tmdbID int) (TVShowDetails, error) {
	var raw rawTVShowDetails
	path := fmt.Sprintf("/tv/%d", tmdbID)
	if err := c.doGet(ctx, path, nil, &raw); err != nil {
		return TVShowDetails{}, err
	}

	seasons := make([]SeasonSummary, 0, len(raw.Seasons))
	for _, s := range raw.Seasons {
		if s.SeasonNumber == 0 {
			continue
		}
		seasons = append(seasons, SeasonSummary{
			SeasonNumber: s.SeasonNumber,
			EpisodeCount: s.EpisodeCount,
			AirDate:      s.AirDate,
		})
	}

	return TVShowDetails{
		TMDBID:       raw.ID,
		Name:         raw.Name,
		FirstAirYear: yearFromDate(raw.FirstAirDate),
		PosterPath:   raw.PosterPath,
		Overview:     raw.Overview,
		Seasons:      seasons,
	}, nil
}

func (c *Client) GetEpisodeDetails(ctx context.Context, tmdbID, season, episode int) (EpisodeDetails, error) {
	var raw rawEpisodeDetails
	path := fmt.Sprintf("/tv/%d/season/%d/episode/%d", tmdbID, season, episode)
	if err := c.doGet(ctx, path, nil, &raw); err != nil {
		return EpisodeDetails{}, err
	}

	return EpisodeDetails{
		RuntimeMinutes: raw.Runtime,
		AirDate:        raw.AirDate,
	}, nil
}

func (c *Client) GetTrending(ctx context.Context) ([]TrendingResult, error) {
	var raw rawTrendingResponse
	if err := c.doGet(ctx, "/trending/all/day", nil, &raw); err != nil {
		return nil, err
	}

	results := make([]TrendingResult, 0, len(raw.Results))
	for _, item := range raw.Results {
		var mt MediaType
		switch item.MediaType {
		case "movie":
			mt = MediaTypeMovie
		case "tv":
			mt = MediaTypeTV
		default:
			continue
		}

		title := item.Title
		releaseDate := item.ReleaseDate
		if mt == MediaTypeTV {
			title = item.Name
			releaseDate = item.FirstAirDate
		}

		results = append(results, TrendingResult{
			TMDBID:      item.ID,
			MediaType:   mt,
			Title:       title,
			ReleaseYear: yearFromDate(releaseDate),
			PosterPath:  item.PosterPath,
			Overview:    item.Overview,
		})
	}

	return results, nil
}
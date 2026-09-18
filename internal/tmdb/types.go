package tmdb

type MediaType string

const (
	MediaTypeMovie MediaType = "movie"
	MediaTypeTV    MediaType = "tv"
)

type SearchResult struct {
	TMDBID      int       `json:"tmdb_id"`
	MediaType   MediaType `json:"media_type"`
	Title       string    `json:"title"`
	ReleaseYear int       `json:"release_year,omitempty"`
	PosterPath  string    `json:"poster_path,omitempty"`
	Overview    string    `json:"overview,omitempty"`
}

type MovieDetails struct {
	TMDBID         int    `json:"tmdb_id"`
	Title          string `json:"title"`
	ReleaseDate    string `json:"release_date,omitempty"`
	RuntimeMinutes int    `json:"runtime_minutes,omitempty"`
	PosterPath     string `json:"poster_path,omitempty"`
	Overview       string `json:"overview,omitempty"`
}


type TVShowDetails struct {
	TMDBID       int             `json:"tmdb_id"`
	Name         string          `json:"name"`
	FirstAirYear int             `json:"first_air_year,omitempty"`
	PosterPath   string          `json:"poster_path,omitempty"`
	Overview     string          `json:"overview,omitempty"`
	Seasons      []SeasonSummary `json:"seasons"`
}


type SeasonSummary struct {
	SeasonNumber int    `json:"season_number"`
	EpisodeCount int    `json:"episode_count"`
	AirDate      string `json:"air_date,omitempty"`
}

type EpisodeDetails struct {
	RuntimeMinutes int    `json:"runtime_minutes,omitempty"`
	AirDate        string `json:"air_date,omitempty"`
}

type TrendingResult struct {
	TMDBID      int       `json:"tmdb_id"`
	MediaType   MediaType `json:"media_type"`
	Title       string    `json:"title"`
	ReleaseYear int       `json:"release_year,omitempty"`
	PosterPath  string    `json:"poster_path,omitempty"`
	Overview    string    `json:"overview,omitempty"`
}

type TrendingResults struct {
	Movies []TrendingResult `json:"movies"`
	Shows  []TrendingResult `json:"shows"`
}
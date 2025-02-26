package services

import (
	"context"
	"encoding/json"
	"fmt"

	customsearch "google.golang.org/api/customsearch/v1"
	"google.golang.org/api/option"
)

// SearchResult represents a single search result from Google Custom Search API
type SearchResult struct {
	Title   string `json:"title"`   // The title of the search result
	Link    string `json:"link"`    // The URL of the search result
	Snippet string `json:"snippet"` // A brief excerpt from the search result
}

// SearchService handles Google Custom Search operations
type SearchService struct {
	apiKey   string // Google API key for authentication
	engineID string // Custom Search Engine ID
}

// NewSearchService creates a new instance of SearchService
// Parameters:
//   - apiKey: Google API key for authentication
//   - engineID: Custom Search Engine ID
//
// Returns:
//   - *SearchService: New instance of search service
func NewSearchService(apiKey, engineID string) *SearchService {
	return &SearchService{
		apiKey:   apiKey,
		engineID: engineID,
	}
}

// Search performs a Google Custom Search and returns structured results
// Parameters:
//   - ctx: Context for handling cancellation and timeouts
//   - query: The search query string
//
// Returns:
//   - []SearchResult: Slice of search results
//   - error: Error if the search fails
func (s *SearchService) Search(ctx context.Context, query string) ([]SearchResult, error) {
	opts := []option.ClientOption{}
	if s.apiKey != "" {
		apiKeyOps := option.WithAPIKey(s.apiKey)
		opts = append(opts, apiKeyOps)
	}
	searchService, err := customsearch.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create search service: %v", err)
	}

	// Configure and execute the search
	search := searchService.Cse.List()
	search.Q(query)
	search.Cx(s.engineID)
	search.Num(5) // Limit results to 5 items

	result, err := search.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to execute search: %v", err)
	}

	// Convert API results to SearchResult structs
	searchResults := make([]SearchResult, 0)
	for _, item := range result.Items {
		searchResults = append(searchResults, SearchResult{
			Title:   item.Title,
			Link:    item.Link,
			Snippet: item.Snippet,
		})
	}

	return searchResults, nil
}

// SearchJSON performs a search and returns results as a JSON string
// Parameters:
//   - ctx: Context for handling cancellation and timeouts
//   - query: The search query string
//
// Returns:
//   - string: JSON string containing search results
//   - error: Error if the search or JSON conversion fails
func (s *SearchService) SearchJSON(ctx context.Context, query string) (string, error) {
	results, err := s.Search(ctx, query)
	if err != nil {
		return "", err
	}

	jsonResult, err := json.Marshal(results)
	if err != nil {
		return "", fmt.Errorf("failed to marshal results: %v", err)
	}

	return string(jsonResult), nil
}

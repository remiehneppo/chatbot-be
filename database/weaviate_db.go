package database

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/tieubaoca/chatbot-be/config"
	"github.com/tieubaoca/chatbot-be/types"
	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/auth"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/filters"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/graphql"
	"github.com/weaviate/weaviate/entities/models"
)

const BATCH_SIZE = 200

var (
	DOCUMENT_CLASS        = "Document"
	DOCUMENT_CLASS_OBJECT = &models.Class{
		Class: DOCUMENT_CLASS,
		Properties: []*models.Property{
			{Name: "content", DataType: []string{"text"}},
			{Name: "title", DataType: []string{"text"}},
			{Name: "source", DataType: []string{"text"}},
			{Name: "tags", DataType: []string{"text[]"}},
			{Name: "custom", DataType: []string{"object"},
				NestedProperties: []*models.NestedProperty{
					{Name: "page", DataType: []string{"text"}},
				},
			},
			{Name: "createdAt", DataType: []string{"int"}},
		},
		VectorIndexType: "hnsw",
	}
)

type WeaviateStore struct {
	client         *weaviate.Client
	text2VecModule string
}

func NewWeaviateStore(config config.WeaviateStoreConfig) (*WeaviateStore, error) {
	var scheme string
	if strings.Contains(config.Host, "https") {
		scheme = "https"
	} else {
		scheme = "http"
	}
	host := strings.TrimPrefix(config.Host, scheme+"://")
	cfg := weaviate.Config{
		Host:   host,
		Scheme: scheme,
	}
	if config.APIKey != "" {
		cfg.AuthConfig = auth.ApiKey{
			Value: config.APIKey,
		}
		cfg.Headers = map[string]string{
			"X-Weaviate-Api-Key":     config.APIKey,
			"X-Weaviate-Cluster-Url": fmt.Sprintf("%s://%s", scheme, host),
		}
	}
	DOCUMENT_CLASS_OBJECT.Vectorizer = config.Text2Vec
	DOCUMENT_CLASS_OBJECT.ModuleConfig = config.ModuleConfig
	client, err := weaviate.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create weaviate client: %v", err)
	}

	schema, err := client.Schema().Getter().Do(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get schema: %v", err)
	}

	hasDocumentClass := false
	for _, class := range schema.Classes {
		if class.Class == DOCUMENT_CLASS {
			hasDocumentClass = true
			break
		}
	}
	// Create Document class if it doesn't exist
	if !hasDocumentClass {
		err = client.Schema().ClassCreator().WithClass(DOCUMENT_CLASS_OBJECT).Do(context.Background())
		if err != nil {
			return nil, fmt.Errorf("failed to create Document class: %v", err)
		}
	}
	return &WeaviateStore{
		client: client,
	}, nil
}

func (s *WeaviateStore) ReInit() error {
	err := s.client.Schema().ClassDeleter().WithClassName(DOCUMENT_CLASS).Do(context.Background())
	if err != nil {
		return fmt.Errorf("failed to delete Document class: %v", err)
	}

	err = s.client.Schema().ClassCreator().WithClass(DOCUMENT_CLASS_OBJECT).Do(context.Background())
	if err != nil {
		return fmt.Errorf("failed to create Document class: %v", err)
	}
	return nil
}

func (s *WeaviateStore) UpsertDocument(ctx context.Context, doc *types.Document, embedding []float32) error {

	// Check if we found any exact matches

	className := DOCUMENT_CLASS
	properties := map[string]interface{}{
		"content":   doc.Content,
		"title":     doc.Metadata.Title,
		"source":    doc.Metadata.Source,
		"tags":      doc.Metadata.Tags,
		"custom":    doc.Metadata.Custom,
		"createdAt": doc.CreatedAt,
	}

	creator := s.client.Data().Creator().
		WithClassName(className).
		WithProperties(properties)

	if embedding != nil {
		creator = creator.WithVector(embedding)
	}

	upsertResult, err := creator.Do(ctx)
	if err != nil {
		return err
	}
	log.Println("UpsertDocument result:", upsertResult.Object.ID)
	return nil
}

func (s *WeaviateStore) BatchInsertDocuments(ctx context.Context, docs []types.Document, embeddings [][]float32) error {
	total := len(docs)
	for i := 0; i < total; i += BATCH_SIZE {
		end := i + BATCH_SIZE
		if end > total {
			end = total
		}

		// Create batch for current chunk
		batcher := s.client.Batch().ObjectsBatcher()

		// Add documents to current batch
		for j := i; j < end; j++ {
			properties := map[string]interface{}{
				"content":   docs[j].Content,
				"title":     docs[j].Metadata.Title,
				"source":    docs[j].Metadata.Source,
				"tags":      docs[j].Metadata.Tags,
				"custom":    docs[j].Metadata.Custom,
				"createdAt": docs[j].CreatedAt,
			}

			// Add embedding if provided
			if embeddings != nil && j < len(embeddings) {
				batcher = batcher.WithObjects(&models.Object{
					Class:      DOCUMENT_CLASS,
					Properties: properties,
					Vector:     embeddings[j],
				})
			} else {
				batcher = batcher.WithObjects(&models.Object{
					Class:      DOCUMENT_CLASS,
					Properties: properties,
				})
			}
		}

		// Execute current batch
		_, err := batcher.Do(ctx)
		if err != nil {
			return fmt.Errorf("failed to insert batch %d-%d: %v", i, end, err)
		}

		log.Printf("Inserted batch %d-%d of %d documents", i, end, total)
	}

	return nil
}

func (s *WeaviateStore) DeleteDocument(ctx context.Context, id string) error {
	return s.client.Data().Deleter().
		WithClassName(DOCUMENT_CLASS).
		WithID(id).
		Do(ctx)
}

func (s *WeaviateStore) AskAI(ctx context.Context, question string, queries []string, metadata types.Metadata, limit int) ([]types.Document, error) {
	fields := []graphql.Field{
		{Name: "content"},
		{Name: "title"},
		{Name: "source"},
		{Name: "tags"},
		{Name: "custom", Fields: []graphql.Field{{Name: "page"}}},
		{Name: "createdAt"},
		{Name: "_additional", Fields: []graphql.Field{{Name: "distance"}, {Name: "id"}}},
	}
	gs := graphql.NewGenerativeSearch().SingleResult(question + "with context {content}")
	response, err := s.client.GraphQL().Get().
		WithClassName(DOCUMENT_CLASS).
		WithFields(
			fields...,
		).
		WithGenerativeSearch(gs).
		WithNearText((&graphql.NearTextArgumentBuilder{}).
			WithConcepts(queries).WithDistance(0.7)).
		WithLimit(limit).
		Do(ctx)

	if err != nil {
		return nil, err
	}
	var docs []types.Document
	if data, ok := response.Data["Get"].(map[string]interface{})[DOCUMENT_CLASS].([]interface{}); ok {
		for _, item := range data {
			if doc, ok := item.(map[string]interface{}); ok {
				document := types.Document{
					Content: doc["content"].(string),
					Metadata: types.Metadata{
						Title:  doc["title"].(string),
						Source: doc["source"].(string),
						Tags:   parseStringArray(doc["tags"]),
						Custom: parseStringMap(doc["custom"]),
					},
					CreatedAt: int64(doc["createdAt"].(float64)),
				}

				docs = append(docs, document)

				if additional, ok := doc["_additional"].(map[string]interface{}); ok {
					document.ID = additional["id"].(string)
					document.Metadata.Custom["distance"] = fmt.Sprintf("%f", additional["distance"].(float64))
					generate := doc["_additional"].(map[string]interface{})["generate"].(map[string]interface{})
					if generate["error"] == nil {
						document.Metadata.Custom["generative"] = generate["singleResult"].(string)
					}
				}
			}
		}
	}
	return docs, nil
}

// Add new method implementation
func (s *WeaviateStore) SearchSimilarWithMetadata(ctx context.Context, queries []string, metadata types.Metadata, limit int) ([]types.Document, []float32, error) {
	fields := []graphql.Field{
		{Name: "content"},
		{Name: "title"},
		{Name: "source"},
		{Name: "tags"},
		{Name: "custom", Fields: []graphql.Field{{Name: "page"}}},
		{Name: "createdAt"},
		{Name: "_additional", Fields: []graphql.Field{{Name: "distance"}, {Name: "id"}}},
	}
	nearVector := s.client.GraphQL().NearTextArgBuilder().
		WithConcepts(queries).
		WithCertainty(0.7)
	// Build where filter for metadata
	where := buildMetadataFilter(metadata)

	// Combined query with both vector similarity and metadata filters
	getBuilder := s.client.GraphQL().Get().
		WithClassName(DOCUMENT_CLASS).
		WithFields(fields...).
		WithNearText(nearVector)
	if limit > 0 {
		getBuilder = getBuilder.WithLimit(limit)
	}
	if where != nil {
		getBuilder = getBuilder.WithWhere(where)
	}

	result, err := getBuilder.Do(ctx)

	if err != nil {
		return nil, nil, err
	}
	if result.Errors != nil {
		return nil, nil, fmt.Errorf("search failed: %v", result.Errors[0].Message)
	}

	// Parse results
	var docs []types.Document
	var distances []float32

	if data, ok := result.Data["Get"].(map[string]interface{})[DOCUMENT_CLASS].([]interface{}); ok {
		for _, item := range data {
			if doc, ok := item.(map[string]interface{}); ok {
				document := types.Document{
					Content: doc["content"].(string),
					Metadata: types.Metadata{
						Title:  doc["title"].(string),
						Source: doc["source"].(string),
						Tags:   parseStringArray(doc["tags"]),
						Custom: parseStringMap(doc["custom"]),
					},
					CreatedAt: int64(doc["createdAt"].(float64)),
				}

				docs = append(docs, document)

				if additional, ok := doc["_additional"].(map[string]interface{}); ok {
					distances = append(distances, float32(additional["distance"].(float64)))
					document.ID = additional["id"].(string)
					document.Metadata.Custom["distance"] = fmt.Sprintf("%f", additional["distance"].(float64))
				}
			}
		}
	}

	return docs, distances, nil
}

// Update SearchSimilar to use common search structure
func (s *WeaviateStore) SearchSimilar(ctx context.Context, queries []string, limit int) ([]types.Document, []float32, error) {
	// Call SearchSimilarWithMetadata with empty metadata
	return s.SearchSimilarWithMetadata(ctx, queries, types.Metadata{}, limit)
}

func (s *WeaviateStore) SearchByMetadata(ctx context.Context, metadata types.Metadata, limit int) ([]types.Document, error) {
	fields := []graphql.Field{
		{Name: "content"},
		{Name: "title"},
		{Name: "source"},
		{Name: "tags"},
		{Name: "custom", Fields: []graphql.Field{{Name: "page"}}},
		{Name: "createdAt"},
		{Name: "_additional", Fields: []graphql.Field{{Name: "distance"}, {Name: "id"}}},
	}

	where := buildMetadataFilter(metadata)

	getBuilder := s.client.GraphQL().Get().
		WithClassName(DOCUMENT_CLASS).
		WithFields(fields...)
	if limit > 0 {
		getBuilder = getBuilder.WithLimit(limit)
	}
	if where != nil {
		getBuilder = getBuilder.WithWhere(where)
	}
	result, err := getBuilder.Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("search failed: %v", err)
	}

	if result.Errors != nil {
		return nil, fmt.Errorf("search failed: %v", result.Errors)
	}

	var docs []types.Document
	if data, ok := result.Data["Get"].(map[string]interface{})[DOCUMENT_CLASS].([]interface{}); ok {
		for _, item := range data {
			if doc, ok := item.(map[string]interface{}); ok {
				document := types.Document{
					ID:      doc["id"].(string),
					Content: doc["content"].(string),
					Metadata: types.Metadata{
						Title:  doc["title"].(string),
						Source: doc["source"].(string),
						Tags:   parseStringArray(doc["tags"]),
						Custom: parseStringMap(doc["custom"]),
					},
					CreatedAt: int64(doc["createdAt"].(float64)),
				}
				docs = append(docs, document)
			}
		}
	}

	return docs, nil
}

func (s *WeaviateStore) CreateCollection(ctx context.Context, name string, dimension int) error {
	classObj := &models.Class{
		Class: name,
		Properties: []*models.Property{
			{Name: "content", DataType: []string{"text"}},
			{Name: "title", DataType: []string{"text"}},
			{Name: "source", DataType: []string{"text"}},
			{Name: "tags", DataType: []string{"text[]"}},
			{Name: "custom", DataType: []string{"object"}},
			{Name: "createdAt", DataType: []string{"int"}},
		},
		Vectorizer: "text2vec-transformers",
		ModuleConfig: map[string]interface{}{
			"text2vec-transformers": map[string]interface{}{
				"model":              "sentence-transformers/all-MiniLM-L6-v2", // default model
				"poolingStrategy":    "masked_mean",
				"vectorizeClassName": false,
			},
		},
		VectorIndexType: "hnsw",
	}

	return s.client.Schema().ClassCreator().WithClass(classObj).Do(ctx)
}

func (s *WeaviateStore) DeleteCollection(ctx context.Context, name string) error {
	return s.client.Schema().ClassDeleter().WithClassName(name).Do(ctx)
}

// Helper functions
func parseStringArray(v interface{}) []string {
	if v == nil {
		return nil
	}
	arr, ok := v.([]interface{})
	if !ok {
		return nil
	}
	result := make([]string, len(arr))
	for i, item := range arr {
		result[i] = item.(string)
	}
	return result
}

func parseStringMap(v interface{}) map[string]string {
	if v == nil {
		return nil
	}
	m, ok := v.(map[string]interface{})
	if !ok {
		return nil
	}
	result := make(map[string]string)
	for k, v := range m {
		result[k] = v.(string)
	}
	return result
}

func buildMetadataFilter(metadata types.Metadata) *filters.WhereBuilder {

	var whereFilter *filters.WhereBuilder

	if metadata.Title != "" {
		whereFilter = filters.Where().WithPath([]string{"title"}).
			WithOperator(filters.Equal).
			WithValueString(metadata.Title)
	}

	if metadata.Source != "" {
		sourceFilter := filters.Where().
			WithPath([]string{"source"}).
			WithOperator(filters.Equal).
			WithValueString(metadata.Source)
		if whereFilter == nil {
			whereFilter = sourceFilter
		} else {
			whereFilter = whereFilter.WithOperator(filters.And).WithOperands([]*filters.WhereBuilder{sourceFilter})
		}

	}

	if len(metadata.Tags) > 0 {
		for _, tag := range metadata.Tags {
			tagFilter := filters.Where().
				WithPath([]string{"tags"}).
				WithOperator(filters.ContainsAny).
				WithValueString(tag)
			if whereFilter == nil {
				whereFilter = tagFilter
			} else {
				whereFilter = whereFilter.WithOperator(filters.And).WithOperands([]*filters.WhereBuilder{tagFilter})
			}
		}
	}

	if len(metadata.Custom) > 0 {
		for key, value := range metadata.Custom {
			customFilter := filters.Where().
				WithPath([]string{"custom", key}).
				WithOperator(filters.Equal).
				WithValueString(value)
			if whereFilter == nil {
				whereFilter = customFilter
			} else {
				whereFilter = whereFilter.WithOperator(filters.And).WithOperands([]*filters.WhereBuilder{customFilter})
			}
		}
	}

	return whereFilter
}

func NewOllamaModuleConfig(apiEndpoint, model, embedModel string) map[string]interface{} {
	return map[string]interface{}{
		"text2vec-ollama": map[string]interface{}{ // Configure the Ollama embedding integration
			"apiEndpoint": apiEndpoint, // Allow Weaviate from within a Docker container to contact your Ollama instance
			"model":       embedModel,  // The model to use
		},
		"generative-ollama": map[string]interface{}{ // Configure the Ollama generative integration
			"apiEndpoint": apiEndpoint, // Allow Weaviate from within a Docker container to contact your Ollama instance
			"model":       model,       // The model to use
		},
	}
}

// Helper function to truncate long strings for logging
func truncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	return s[:maxLength] + "..."
}

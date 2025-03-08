package types

// Document represents a knowledge base document
type Document struct {
	ID        string   `bson:"_id" json:"id"`
	Content   string   `bson:"content" json:"content"`
	Metadata  Metadata `bson:"metadata" json:"metadata"`
	CreatedAt int64    `bson:"created_at" json:"created_at"`
}

// Metadata contains additional document information
type Metadata struct {
	Title  string            `bson:"title" json:"title"`
	Source string            `bson:"source" json:"source"`
	Tags   []string          `bson:"tags" json:"tags"`
	Custom map[string]string `bson:"custom" json:"custom"`
}

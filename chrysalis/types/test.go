package types

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// IndicativeChain represents a document in the IndicativeChains collection.
type IndicativeChain struct {
	SourcePaths []string `json:"source_paths,omitempty"`
}

// Test represents a document in the Tests collection.
type Test struct {
	ID               bson.ObjectID     `bson:"_id" json:"_id"`
	Key              bson.ObjectID     `bson:"key" json:"key"`
	Path             string            `bson:"path" json:"path"`
	IndicativeChains []IndicativeChain `bson:"indicativeChains" json:"indicativeChains"`
	UpdatedAt        time.Time         `bson:"updatedAt" json:"updatedAt"`
	ExpireAt         time.Time         `bson:"expireAt" json:"expireAt"`
}

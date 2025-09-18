package types

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// IndicativeChain represents a document in the IndicativeChains collection.
type IndicativeChain struct {
	SourcePaths []string `json:"source_paths,omitempty"`
}

// Test represents a document in the Tests collection.
type Test struct {
	ID               primitive.ObjectID `bson:"_id" json:"_id"`
	Key              primitive.ObjectID `bson:"key" json:"key"`
	Path             string             `bson:"path" json:"path"`
	IndicativeChains []IndicativeChain  `bson:"indicativeChains" json:"indicativeChains"`
	UpdatedAt        time.Time          `bson:"updatedAt" json:"updatedAt"`
	ExpireAt         time.Time          `bson:"expireAt" json:"expireAt"`
}

package types

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Identifier represents a document in the Identifiers collection. The ID in this is used to identify tests, chains
type Identifier struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	AccountID string             `bson:"accountId" json:"accountId"`
	OrgID     string             `bson:"orgId" json:"orgId"`
	ProjectID string             `bson:"projectId" json:"projectId"`
	Repo      string             `bson:"repo" json:"repo"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	ExpiresAt time.Time          `bson:"expiresAt" json:"expiresAt"`
}

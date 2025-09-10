package types

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Test state constants for representing different test execution outcomes
const (
	SUCCESS TestState = "SUCCESS"
	FAILURE TestState = "FAILURE"
	FLAKY   TestState = "FLAKY"
	UNKNOWN TestState = "UNKNOWN"
)

// Chain represents a document in the Chains collection with state field.
type Chain struct {
	ID           primitive.ObjectID `bson:"_id" json:"_id"`
	Key          primitive.ObjectID `bson:"key" json:"key"`
	Path         string             `bson:"path" json:"path"`
	TestChecksum string             `bson:"testChecksum" json:"testChecksum"`
	Checksum     string             `bson:"checksum" json:"checksum"`
	State        TestState          `bson:"state" json:"state"`
	ExtraInfo    map[string]string  `bson:"extraInfo" json:"extraInfo"`
	UpdatedAt    time.Time          `bson:"updatedAt" json:"updatedAt"`
	ExpireAt     time.Time          `bson:"expireAt" json:"expireAt"`
}

type TestState string

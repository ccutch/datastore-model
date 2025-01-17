package db

import (
	"context"

	"google.golang.org/appengine/datastore"
)

type KeyResolver struct {
	context context.Context
}

// NewKeyResolver creates a new instance of *KeyResolver
func NewKeyResolver(c context.Context) *KeyResolver {
	return &KeyResolver{
		context: c,
	}
}

// Resolve resolves the datastore key for the given entity
// by either assembling it based on the structs tags
// or by creating an auto generated key in case no tags are
// provided
//
// ErrMissingStringId is returned in case a string field
// is tagged with db:"id" and is empty
//
// ErrMissingIntId is returned in case an int field
// is tagged with db:"id" and is 0
func (this *KeyResolver) Resolve(e Entity) (*Metadata, error) {
	metadata := &Metadata{}

	if err := NewKeyResolverExtractorChain(metadata).ExtractFrom(e); err != nil {
		return nil, err
	}

	if metadata.IntID != 0 && metadata.StringID != "" {
		return nil, ErrMultipleIdFields
	}

	k := datastore.NewKey(
		this.context,
		metadata.Kind,
		metadata.StringID,
		metadata.IntID,
		metadata.Parent,
	)

	e.SetKey(k)

	if metadata.CacheStringID == "" {
		metadata.CacheStringID = e.StringId()
	}

	return metadata, nil
}

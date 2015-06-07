package db

import (
	"appengine"
	"appengine/datastore"
	"time"
)

type Clock func() time.Time

type entity interface {
	HasKey() bool
	Key() *datastore.Key
	SetKey(*datastore.Key)
	Parent() *datastore.Key
	SetParent(*datastore.Key)
	SetCreatedAt(time.Time)
}

type resource interface {
	//	Url() string
	StringId() string
	SetStringId(string) error
}

type model interface {
	entity
	resource
}

type Datasource interface {
	Create(entity) error
	CreateAll(...entity) error
	Update(entity) error
	Load(entity) error
	Delete(entity) error
	DeleteAll(...entity) error
	Query(*Query) *Querier
}

// Datastore Service that provides a set of
// operations to make it easy on you when
// working with appengine datastore
//
// It works along with db.Model in order to
// provide its features.
//
type Datastore struct {
	context appengine.Context
	Clock   Clock
	KeyResolver *KeyResolver
}

func NewDatastore(c appengine.Context) Datastore {
	return Datastore{
		context: c,
		Clock: time.Now,
		KeyResolver: NewKeyResolver(c),
	}
}

// Create creates a new entity in datastore
// using the key generated by the keyProvider
func (this Datastore) Create(e entity) error {
	if err := this.AssignNewKey(e); err != nil {
		return err
	}
	e.SetCreatedAt(this.Clock())
	_, err := datastore.Put(this.context, e.Key(), e)
	return err
}

// CreateAll creates entities in batch
func (this Datastore) CreateAll(es ...entity) error {
	keys := make([]*datastore.Key, len(es))
	for i, e := range es {
		if err := this.AssignNewKey(e); err != nil {
			// rollback changes to created at of previous entities
			for j := i; j >= 0; j-- {
				es[j].SetCreatedAt(time.Time{})
			}
			return err
		}
		keys[i] = e.Key()
		e.SetCreatedAt(this.Clock())
	}
	_, err := datastore.PutMulti(this.context, keys, es)
	return err
}

// Update updated an entity in datastore
func (this Datastore) Update(e entity) error {
	if err := this.ResolveEntityKey(e); err != nil {
		return err
	}
	_, err := datastore.Put(this.context, e.Key(), e)
	return err
}

func (this Datastore) UpdateAll(es ...entity) error {
	keys := make([]*datastore.Key, len(es))
	for i, e := range es {
		if err := this.ResolveEntityKey(e); err != nil {
			return err
		}
		keys[i] = e.Key()
	}
	_, err := datastore.PutMulti(this.context, keys, es)
	return err
}

// Load loads entity data from datastore
func (this Datastore) Load(e entity) error {
	if err := this.ResolveEntityKey(e); err != nil {
		return err
	}
	return datastore.Get(this.context, e.Key(), e)
}

func (this Datastore) LoadAll(es ...entity) error {
	keys := make([]*datastore.Key, len(es))
	for i, e := range es {
		if err := this.ResolveEntityKey(e); err != nil {
			return err
		}
		keys[i] = e.Key()
	}
	return datastore.GetMulti(this.context, keys, es)
}

// Delete deletes an entity from datastore
func (this Datastore) Delete(e entity) error {
	if err := this.ResolveEntityKey(e); err != nil {
		return err
	}
	return datastore.Delete(this.context, e.Key())
}

func (this Datastore) DeleteAll(es ...entity) error {
	keys := make([]*datastore.Key, len(es))
	for i, e := range es {
		if err := this.ResolveEntityKey(e); err != nil {
			return err
		}
		keys[i] = e.Key()
	}
	return datastore.DeleteMulti(this.context, keys)
}

// Query returns an instance of Querier
func (this Datastore) Query(q *Query) *Querier {
	return &Querier{this.context, q}
}

// AssignEntityKey generates a new datastore key for the given entity
//
// The Key components are derived from the entity struct through reflection
// Fields tagged with `db:"id"` are used in the key as a StringID if
// the field type is string, or IntID in case its type is any int type
//
// In case multiple fields are tagged with `db:"id"`, the first field
// is selected to be used as id in the key
//
// If no field is tagged, the key is generated using the default values
// for StringID and IntID, causing the key to be auto generated
func (this Datastore) AssignNewKey(e entity) error {
	return this.KeyResolver.Resolve(e)
}

// ResolveEntityKey assembles the key for the given entity based
// on its struct tags
//
// ErrMissingAutoGeneratedKey is returned in case the struct
// has no db:"id" tag,
func (this Datastore) ResolveEntityKey(e entity) error {
	if err := this.KeyResolver.Resolve(e); err != nil {
		return err
	}

	if this.KeyResolver.IsAutoGenerated() {
		e.SetKey((*datastore.Key)(nil))
		return ErrMissingAutoGeneratedKey
	}
	return nil
}

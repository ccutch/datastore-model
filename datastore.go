package db

import (
	"appengine"
	"appengine/datastore"
	"errors"
)

var (
	ErrNoSuchEntity = errors.New("Entity not found")
	ErrEntityExists = errors.New("Entity already exists")
)

type entity interface {
	Kind() string
	HasKey() bool
	Key() *datastore.Key
	setKey(*datastore.Key)
	NewKey(appengine.Context) *datastore.Key
	UUID() string
	SetUUID(uuid string) error
}

// Datastore Service that provides a set of
// operations to make it easy on you when
// working with appengine datastore
//
// It works along with db.Model in order to
// provide its features.
type Datastore struct {
	Context appengine.Context
}

// Create creates a new entity in datastore
// using the key generated by the keyProvider
//
// ErrEntityExists is returned in case the key
// generated by the KeyProvider is already being
// used
func (this Datastore) Create(e entity) error {
	if err := this.Load(e); err == nil {
		return ErrEntityExists
	}

	key, err := datastore.Put(this.Context, e.NewKey(this.Context), e);
	e.setKey(key)
	return err
}

func (this Datastore) Save(e entity) error {
    if err := this.Load(e); err != nil {
        return err
    }
    _, err := datastore.Put(this.Context, e.Key(), e);
    return err
}

// Load loads entity data from datastore
//
// In case the entity has no key yet assigned
// a new one is created by the entity itself
// and used to retrieve the entity data from
// datastore
//
// ErrNoSuchEntity is returned in case no
// entity is found for the given key
func (this Datastore) Load(e entity) error {
	if !e.HasKey() {
		e.setKey(e.NewKey(this.Context))
	}
	return datastore.Get(this.Context, e.Key(), e)
}

// Delete deletes an entity from datastore
//
// ErrNoSuchEntity is returned in case the
// key provided does not match any existent
// entity
func (this Datastore) Delete(e entity) error {
	if err := this.Load(e); err == datastore.ErrNoSuchEntity {
		return ErrNoSuchEntity
	}

	if !e.HasKey() {
		e.setKey(e.NewKey(this.Context))
	}

	return datastore.Delete(this.Context, e.Key())
}
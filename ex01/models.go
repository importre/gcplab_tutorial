package ex01

import (
	"appengine/datastore"
)

type Jange struct {
	IdStr    string
	Password string
	EncKey   *datastore.Key
}

type Nakseo struct {
	EncKey      *datastore.Key `json:"encKey"`
	Content     string         `json:"content"`
	Owner       string         `json:"owner"`
	EncOwnerKey *datastore.Key `json:"encOwnerKey"`
	RegDate     string         `json:"regDate"`
}

type NakseoResult struct {
	Result    string   `json:"result"`
	Nakseo    []*Nakseo `json:"nakseo"`
	Index     string   `json:"idx"`
	NextToken int      `json:"next"`
}

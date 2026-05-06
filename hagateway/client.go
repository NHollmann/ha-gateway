package hagateway

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

type Entity struct {
	Name     string
	CanWrite bool `mapstructure:"can_write"`
}

func (e *Entity) MatchName(name string) bool {
	if e.Name == "*" {
		return true
	}

	if strings.HasSuffix(e.Name, "*") && strings.HasPrefix(name, e.Name[:len(e.Name)-2]) {
		return true
	}

	return e.Name == name
}

type Client struct {
	Name      string
	TokenHash string `mapstructure:"token_hash"`
	CanWrite  bool   `mapstructure:"can_write"`
	Entities  []Entity
}

func (c *Client) CanReadEntity(entityName string) bool {
	for i := range c.Entities {
		if c.Entities[i].MatchName(entityName) {
			return true
		}
	}
	return false
}

func (c *Client) CanWriteEntity(entityName string) bool {
	for i := range c.Entities {
		if c.Entities[i].MatchName(entityName) {
			return c.CanWrite || c.Entities[i].CanWrite
		}
	}
	return false
}

type Clients struct {
	authAccess map[string]*Client
}

func (cs *Clients) Add(c *Client) bool {
	if c == nil {
		return false
	}
	if cs.authAccess == nil {
		cs.authAccess = map[string]*Client{}
	}
	if _, ok := cs.authAccess[strings.ToLower(c.TokenHash)]; ok {
		return false
	}
	cs.authAccess[strings.ToLower(c.TokenHash)] = c
	return true
}

func (cs *Clients) FindByToken(token string) *Client {
	tokenHash := fmt.Sprintf("%x", sha256.Sum256([]byte(token)))
	client, ok := cs.authAccess[tokenHash]
	if !ok {
		return nil
	}
	return client
}

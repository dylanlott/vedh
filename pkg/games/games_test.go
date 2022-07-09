package games

import (
	"testing"

	"github.com/matryer/is"
)

func TestInMemoryGame(t *testing.T) {
	is := is.New(t)

	m := &MemStore{
		games: make(map[string]Game),
	}
	game := &FullGame{
		id: "0xACAB",
	}

	created, err := m.Create(game)
	is.NoErr(err)
	t.Logf("Created: %v", created)

}

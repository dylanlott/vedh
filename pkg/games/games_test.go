package games

import (
	"testing"
	"time"

	"github.com/matryer/is"
)

func TestInMemoryGame(t *testing.T) {
	is := is.New(t)

	t.Run("should create a game", func(t *testing.T) {
		m := &MemStore{
			games: make(map[string]Game),
		}

		now := time.Now()
		created, err := m.NewFullGame("test", []Player{
			&player{boardstate: make(JSON, 0)},
		})
		is.NoErr(err)
		is.True(len(m.games) > 0)
		is.True(created.createdAt.After(now))
	})

	t.Run("should return a game", func(t *testing.T) {
		m := &MemStore{
			games: map[string]Game{
				"foo": &FullGame{
					id:        "foo",
					players:   []Player{},
					createdAt: time.Time{},
				},
			},
		}

		games, err := m.List()
		is.NoErr(err)
		is.True(len(games) == 1)
	})
}

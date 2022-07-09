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
		created, err := m.NewFullGame("test", []Player{})
		is.NoErr(err)
		is.True(len(m.games) > 0)
		is.True(created.createdAt.After(now))
	})

	t.Run("should return a list of game", func(t *testing.T) {
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

	t.Run("should update a player's boardstate", func(t *testing.T) {
		m := &MemStore{
			games: map[string]Game{
				"foo": &FullGame{
					id: "foo",
					players: []Player{
						&inMemPlayer{
							state: map[string]interface{}{
								"id": "bar",
							},
						},
					},
					createdAt: time.Time{},
				},
			},
		}

		// get game foo
		g, err := m.Get("foo")
		is.NoErr(err)

		// get player bar
		p, err := g.Get("bar")
		is.NoErr(err)

		// get player's boardstate
		bs, err := p.Boardstate()
		is.NoErr(err)

		// mutate a piece of state
		bs["biz"] = "baz"

		// update player
		err = p.Sync(bs)
		is.NoErr(err)

		got, err := p.Boardstate()
		is.NoErr(err)
		is.Equal(got["biz"], "baz")
		// TODO: assert that events are emitted through PubSub
	})
}

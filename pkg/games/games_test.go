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

		// attach a game so that publishing works
		err = p.AttachGame(g)
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
	})

	t.Run("should join a game", func(t *testing.T) {
		is := is.New(t)
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

		game, err := m.Join("foo", &inMemPlayer{
			state: map[string]interface{}{
				"id": "baz",
			},
		})
		is.NoErr(err)

		players, err := game.Players()
		is.NoErr(err)
		is.Equal(len(players), 2)
	})

	t.Run("should emit a join game event", func(t *testing.T) {
		is := is.New(t)
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

		want, err := m.Get("foo")
		is.NoErr(err)

		gamechan, err := want.Subscribe()
		is.NoErr(err)

		_, err = want.Join(&inMemPlayer{
			state: map[string]interface{}{
				"id": "biz",
			},
		})
		is.NoErr(err)

		got := <-gamechan
		is.Equal(want.ID(), got.ID())
	})

	t.Run("should attach a Game to a Player", func(t *testing.T) {
		is := is.New(t)
		m := &MemStore{
			games: map[string]Game{},
		}

		game, err := m.NewFullGame("fuzz", nil)
		is.NoErr(err)
		is.Equal(game.ID(), "fuzz")
		is.Equal(len(game.players), 0)

		player, err := NewPlayer("fizz")
		is.NoErr(err)
		is.Equal(player.ID(), "fizz")

		got, err := game.Join(player)
		is.NoErr(err)
		is.True(got.ID() == game.ID())
	})

	t.Run("should emit a game event on player sync", func(t *testing.T) {
		is := is.New(t)
		m := &MemStore{
			games: map[string]Game{},
		}

		g, err := m.NewFullGame("fuzz", nil)
		is.NoErr(err)
		is.Equal(g.ID(), "fuzz")
		is.Equal(len(g.players), 0)

		player, err := NewPlayer("fizz")
		is.NoErr(err)
		is.Equal(player.ID(), "fizz")

		game, err := g.Join(player)
		is.NoErr(err)
		is.True(game.ID() == g.ID())

		sub, err := game.Subscribe()
		is.NoErr(err)

		fetchedPlayer, err := game.Get("fizz")
		is.NoErr(err)

		syncErr := fetchedPlayer.Sync(JSON{
			"foo": "bar",
		})
		is.NoErr(syncErr)

		got := <-sub
		is.Equal(got.ID(), game.ID())
	})
}

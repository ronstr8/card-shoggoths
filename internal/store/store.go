package store

import "card-shoggoths/internal/game"

type GameStore interface {
	Save(id string, state *game.GameState) error
	Load(id string) (*game.GameState, error)
}

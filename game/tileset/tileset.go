package tileset

import (
	"math/rand"
)

type Tile struct {
	Number uint
	Letter string
}

type Tileset struct {
	tiles []Tile
}

func New() *Tileset {
	tileset := Tileset{}
	letters := [9]string{"A", "B", "C", "D", "E", "F", "G", "H", "I"}
	for number := 1; number < 13; number++ {
		for _, letter := range letters {
			tileset.tiles = append(tileset.tiles, Tile{uint(number), letter})
		}
	}

	return &tileset
}

// Extracts a random tile from the tileset and returns it
func (t *Tileset) Draw() Tile {
	pos := rand.Intn(len(t.tiles) - 1)
	tile := t.tiles[pos]
	t.tiles = append(t.tiles[:pos], t.tiles[pos+1:]...)
	return tile
}
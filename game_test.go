package game

import (
	"github.com/svera/acquire/board"
	"github.com/svera/acquire/corporation"
	"github.com/svera/acquire/fsm"
	"github.com/svera/acquire/player"
	"github.com/svera/acquire/tile"
	"github.com/svera/acquire/tileset"
	"testing"
)

func TestNewGameWrongNumberPlayers(t *testing.T) {
	players, corporations, board, tileset := setup()
	players = players[:1]

	if _, err := New(board, players, corporations, tileset); err.Error() != WrongNumberPlayers {
		t.Errorf("Game must not be created with less than 3 players, got %d", len(players))
	}
}

func TestNewGameNotUniqueCorpNames(t *testing.T) {
	players, corporations, board, tileset := setup()

	corporations[0] = corporation.NewStub("A", 0)
	corporations[1] = corporation.NewStub("A", 0)

	if _, err := New(board, players, corporations, tileset); err.Error() != CorpNamesNotUnique {
		t.Errorf("Corporations must have unique names, expecting %s error, got %s", CorpNamesNotUnique, err.Error())
	}
}

func TestNewGameWrongNumberOfCorpsPerClass(t *testing.T) {
	players, corporations, board, tileset := setup()

	corporations[2] = corporation.NewStub("C", 0)

	if _, err := New(board, players, corporations, tileset); err.Error() != WrongNumberCorpsClass {
		t.Errorf("Game must catch wrong number of corporations per class")
	}
}

func TestNewGameInitsPlayersTilesets(t *testing.T) {
	players, corporations, board, tileset := setup()
	New(board, players, corporations, tileset)

	for i, player := range players {
		if len(player.Tiles()) != 6 {
			t.Errorf("Players must have 6 tiles at the beginning, player %d got %d", i, len(player.Tiles()))
		}
	}
}

func TestAreEndConditionsReached(t *testing.T) {
	players, corporations, board, tileset := setup()
	game, _ := New(board, players, corporations, tileset)

	if game.AreEndConditionsReached() {
		t.Errorf("End game conditions not reached (no active corporations) but detected as it were")
	}

	corporations[0].(*corporation.Stub).SetSize(41)

	if !game.AreEndConditionsReached() {
		t.Errorf("End game conditions reached (a corporation bigger than 40 tiles) but not detected")
	}

	corporations[0].(*corporation.Stub).SetSize(11)

	if !game.AreEndConditionsReached() {
		t.Errorf("End game conditions reached (all active corporations safe) but not detected")
	}

	corporations[0].(*corporation.Stub).SetSize(11)
	corporations[1].(*corporation.Stub).SetSize(2)

	if game.AreEndConditionsReached() {
		t.Errorf("End game conditions not reached but detected as it were")
	}

}

func TestPlayTileFoundCorporation(t *testing.T) {
	players, corporations, bd, ts := setup()
	tileToPlay := tile.New(6, "E", tile.Orphan{})
	bd.PutTile(tile.New(5, "E", tile.Orphan{}))

	game, _ := New(bd, players, corporations, ts)
	playerTiles := players[0].Tiles()
	players[0].
		DiscardTile(playerTiles[0]).
		PickTile(tileToPlay)
	game.PlayTile(tileToPlay)

	if game.state.Name() != "FoundCorp" {
		t.Errorf("Game must be in state FoundCorp, got %s", game.state.Name())
	}
}

func TestFoundCorporation(t *testing.T) {
	players, corporations, bd, ts := setup()
	game, _ := New(bd, players, corporations, ts)
	if err := game.FoundCorporation(corporations[0]); err == nil {
		t.Errorf("Game in a state different than FoundCorp must not execute FoundCorporation()")
	}
	game.state = &fsm.FoundCorp{}
	newCorpTiles := []tile.Interface{
		tile.New(5, "E", tile.Orphan{}),
		tile.New(6, "E", tile.Orphan{}),
	}
	game.newCorpTiles = newCorpTiles
	game.FoundCorporation(corporations[0])
	if game.state.Name() != "BuyStock" {
		t.Errorf("Game must be in state BuyStock, got %s", game.state.Name())
	}
	if players[0].Shares(corporations[0]) != 1 {
		t.Errorf("Player must have 1 share of corporation stock, got %d", players[0].Shares(corporations[0]))
	}
	if corporations[0].Size() != 2 {
		t.Errorf("Corporation must have 2 tiles, got %d", corporations[0].Size())
	}
	if game.board.Cell(newCorpTiles[0].Number(), newCorpTiles[0].Letter()).Owner() != corporations[0] || game.board.Cell(newCorpTiles[1].Number(), newCorpTiles[1].Letter()).Owner() != corporations[0] {
		t.Errorf("Corporation tiles are not set on board")
	}
}

func TestPlayTileGrowCorporation(t *testing.T) {
	players, corporations, bd, ts := setup()
	tileToPlay := tile.New(6, "E", tile.Orphan{})
	corpTiles := []tile.Interface{
		tile.New(7, "E", corporations[0]),
		tile.New(8, "E", corporations[0]),
	}
	corporations[0].Grow(len(corpTiles))
	bd.SetTiles(corporations[0], corpTiles)
	bd.PutTile(tile.New(5, "E", tile.Orphan{}))

	game, _ := New(bd, players, corporations, ts)
	playerTiles := players[0].Tiles()
	players[0].
		DiscardTile(playerTiles[0]).
		PickTile(tileToPlay)
	game.PlayTile(tileToPlay)

	expectedCorpSize := 4

	if game.state.Name() != "BuyStock" {
		t.Errorf("Game must be in state BuyStock, got %s", game.state.Name())
	}
	if corporations[0].Size() != expectedCorpSize {
		t.Errorf("Corporation size must be %d, got %d", expectedCorpSize, corporations[0].Size())
	}
}

// Testing this merge:
//   4 5 6 7 8 9
// E [][]><[][][]
//
// In this case, players 0 and 1 are the majority shareholders and player 2 is the minority one
// of corporation 0, with a size of 2
func TestPlayTileMergeCorporationsMultipleMajorityShareholders(t *testing.T) {
	players, corporations, bd, ts := setup()
	setupPlayTileMerge(corporations, bd)
	tileToPlay := tile.New(6, "E", tile.Orphan{})

	game, _ := New(bd, players, corporations, ts)
	playerTiles := players[0].Tiles()
	players[0].
		DiscardTile(playerTiles[0]).
		PickTile(tileToPlay)
	players[0].(*player.Stub).SetShares(corporations[0], 6)
	players[1].(*player.Stub).SetShares(corporations[0], 6)
	players[2].(*player.Stub).SetShares(corporations[0], 4)

	game.PlayTile(tileToPlay)
	expectedPlayer0Cash := 7500
	if players[0].Cash() != expectedPlayer0Cash {
		t.Errorf("Player havent received the correct bonus, must have %d$, got %d$", expectedPlayer0Cash, players[0].Cash())
	}
	expectedPlayer1Cash := 7500
	if players[1].Cash() != expectedPlayer1Cash {
		t.Errorf("Player havent received the correct bonus, must have %d$, got %d$", expectedPlayer1Cash, players[1].Cash())
	}
	expectedPlayer2Cash := 6000
	if players[2].Cash() != expectedPlayer2Cash {
		t.Errorf("Player havent received the correct bonus, must have %d$, got %d$", expectedPlayer2Cash, players[2].Cash())
	}
}

// Testing this merge:
//   4 5 6 7 8 9
// E [][]><[][][]
//
// In this case, player 0 is the majority shareholder and players 1 and 2 are the minority ones
// of corporation 0, with a size of 2
func TestPlayTileMergeCorporationsMultipleMinorityhareholders(t *testing.T) {
	players, corporations, bd, ts := setup()
	setupPlayTileMerge(corporations, bd)
	tileToPlay := tile.New(6, "E", tile.Orphan{})

	game, _ := New(bd, players, corporations, ts)
	playerTiles := players[0].Tiles()
	players[0].
		DiscardTile(playerTiles[0]).
		PickTile(tileToPlay)
	players[0].(*player.Stub).SetShares(corporations[0], 6)
	players[1].(*player.Stub).SetShares(corporations[0], 4)
	players[2].(*player.Stub).SetShares(corporations[0], 4)

	game.PlayTile(tileToPlay)
	expectedPlayer0Cash := 8000
	if players[0].Cash() != expectedPlayer0Cash {
		t.Errorf("Player havent received the correct bonus, must have %d$, got %d$", expectedPlayer0Cash, players[0].Cash())
	}
	expectedPlayer1Cash := 6500
	if players[1].Cash() != expectedPlayer1Cash {
		t.Errorf("Player havent received the correct bonus, must have %d$, got %d$", expectedPlayer1Cash, players[1].Cash())
	}
	expectedPlayer2Cash := 6500
	if players[2].Cash() != expectedPlayer2Cash {
		t.Errorf("Player havent received the correct bonus, must have %d$, got %d$", expectedPlayer2Cash, players[2].Cash())
	}
}

// Testing this merge:
//   4 5 6 7 8 9
// E [][]><[][][]
//
// In this case, only player 0 has shares of the defunct corp 0, which has
// a size of 2, thus getting both majority and minority bonuses
func TestPlayTileMergeCorporationsOneShareholder(t *testing.T) {
	players, corporations, bd, ts := setup()
	setupPlayTileMerge(corporations, bd)

	tileToPlay := tile.New(6, "E", tile.Orphan{})

	game, _ := New(bd, players, corporations, ts)
	playerTiles := players[0].Tiles()
	players[0].
		DiscardTile(playerTiles[0]).
		PickTile(tileToPlay)
	players[0].(*player.Stub).SetShares(corporations[0], 6)

	game.PlayTile(tileToPlay)
	expectedPlayerCash := 9000
	if players[0].Cash() != expectedPlayerCash {
		t.Errorf("Player havent received the correct bonus, must have %d$, got %d$", expectedPlayerCash, players[0].Cash())
	}
}

// Set ups the board this way for merge tests
//   4 5 6 7 8 9
// E [][]><[][][]
func setupPlayTileMerge(corporations [7]corporation.Interface, bd board.Interface) {
	corp0Tiles := []tile.Interface{
		tile.New(4, "E", corporations[0]),
		tile.New(5, "E", corporations[0]),
	}
	corp1Tiles := []tile.Interface{
		tile.New(7, "E", corporations[1]),
		tile.New(8, "E", corporations[1]),
		tile.New(9, "E", corporations[1]),
	}
	corporations[0].Grow(len(corp0Tiles))
	corporations[1].Grow(len(corp1Tiles))
	bd.SetTiles(corporations[0], corp0Tiles)
	bd.SetTiles(corporations[1], corp1Tiles)
}

func TestBuyStock(t *testing.T) {
	players, corporations, bd, ts := setup()
	corporations[0].Grow(2)
	buys := map[int]int{0: 2}
	var expectedAvailableStock int = 23
	var expectedPlayerStock int = 2
	game, _ := New(bd, players, corporations, ts)
	game.state = &fsm.BuyStock{}
	game.BuyStock(buys)

	if corporations[0].Stock() != expectedAvailableStock {
		t.Errorf("Corporation stock shares have not decreased, must be %d, got %d", expectedAvailableStock, corporations[0].Stock())
	}
	if players[0].Shares(corporations[0]) != expectedPlayerStock {
		t.Errorf("Player stock shares have not increased, must be %d, got %d", expectedPlayerStock, players[0].Shares(corporations[0]))
	}
}

func TestBuyStockWithNotEnoughCash(t *testing.T) {
	players, corporations, bd, ts := setup()
	players[0].(*player.Stub).SetCash(100)

	corporations[0].Grow(2)

	buys := map[int]int{0: 2}
	game, _ := New(bd, players, corporations, ts)
	err := game.BuyStock(buys)
	if err == nil {
		t.Errorf("Trying to buy stock shares without enough money must throw error")
	}
}

// Testing that if player has an permanently unplayable tile, this is exchanged:
// In the following example, tile 6D is unplayable because it would merge safe
// corporations 0 and 1
//
//    5 6  7 8
// D [0]><[1]
func TestDrawTile(t *testing.T) {
	players, corporations, bd, ts := setup()
	corporations[0].(*corporation.Stub).SetSize(11)
	corporations[1].(*corporation.Stub).SetSize(15)
	unplayableTile := tile.New(6, "D", tile.Orphan{})
	bd.SetTiles(corporations[0], []tile.Interface{tile.New(5, "D", tile.Orphan{})})
	bd.SetTiles(corporations[1], []tile.Interface{tile.New(7, "D", tile.Orphan{})})

	game, _ := New(bd, players, corporations, ts)
	players[0].(*player.Stub).SetTiles([]tile.Interface{unplayableTile})
	game.tileset.(*tileset.Stub).DiscardTile(unplayableTile)
	game.state = &fsm.BuyStock{}
	game.drawTile()
	for _, tile := range players[0].Tiles() {
		if tile.Number() == unplayableTile.Number() && tile.Letter() == unplayableTile.Letter() {
			t.Errorf("Unplayable tile not discarded after drawing new tile, got %v", players[0].Tiles())
		}
	}
}

func setup() ([]player.Interface, [7]corporation.Interface, board.Interface, tileset.Interface) {
	var players []player.Interface
	players = append(players, player.NewStub("Test1"))
	players = append(players, player.NewStub("Test2"))
	players = append(players, player.NewStub("Test3"))

	var corporations [7]corporation.Interface
	corporations[0] = corporation.NewStub("A", 0)
	corporations[1] = corporation.NewStub("B", 0)
	corporations[2] = corporation.NewStub("C", 1)
	corporations[3] = corporation.NewStub("D", 1)
	corporations[4] = corporation.NewStub("E", 1)
	corporations[5] = corporation.NewStub("F", 2)
	corporations[6] = corporation.NewStub("G", 2)

	board := board.New()
	tileset := tileset.NewStub()
	return players, corporations, board, tileset
}

func slicesSameContent(slice1 []player.Interface, slice2 []player.Interface) bool {
	if len(slice1) != len(slice2) {
		return false
	}
	var inSlice bool
	for _, val1 := range slice1 {
		inSlice = false
		for _, val2 := range slice2 {
			if val1 == val2 {
				inSlice = true
				break
			}
		}
		if !inSlice {
			return false
		}
	}
	return true
}

package board

import (
	"github.com/svera/acquire/corporation"
	"reflect"
	//"sort"
	"github.com/svera/acquire/tile"
	"testing"
)

func TestPutTile(t *testing.T) {
	board := New()
	tile := tile.New(5, "B", tile.Unincorporated{})
	board.PutTile(tile)
	if board.grid[5]["B"].Owner().Type() != "unincorporated" {
		t.Errorf("Position %d%s was not put on the board", 5, "B")
	}
}

func TestTileFoundCorporation(t *testing.T) {
	board := New()
	board.grid[5]["D"] = tile.New(5, "D", tile.Unincorporated{})
	board.grid[6]["C"] = tile.New(6, "C", tile.Unincorporated{})
	board.grid[6]["E"] = tile.New(6, "E", tile.Unincorporated{})
	board.grid[7]["D"] = tile.New(7, "D", tile.Unincorporated{})
	foundingTile := tile.New(6, "D", tile.Unincorporated{})
	found, corporationTiles := board.TileFoundCorporation(
		foundingTile,
	)

	expectedCorporationTiles := []tile.Interface{
		foundingTile,
		board.grid[5]["D"],
		board.grid[6]["C"],
		board.grid[6]["E"],
		board.grid[7]["D"],
	}

	if !found {
		t.Errorf("TileFoundCorporation() must return true")
	}
	if !slicesSameCells(corporationTiles, expectedCorporationTiles) {
		t.Errorf("Position %d%s must found a corporation with tiles %v, got %v instead", 6, "D", expectedCorporationTiles, corporationTiles)
	}
}

func TestTileNotFoundCorporation(t *testing.T) {
	board := New()
	corp, _ := corporation.New("Test 1", 0)
	found, corporationTiles := board.TileFoundCorporation(tile.New(6, "D", tile.Unincorporated{}))
	if found {
		t.Errorf("Position %d%s must not found a corporation, got %v instead", 6, "D", corporationTiles)
	}

	board.grid[5]["E"] = tile.New(5, "E", tile.Unincorporated{})
	board.grid[7]["E"] = tile.New(5, "E", corp)
	board.grid[6]["D"] = tile.New(6, "D", tile.Unincorporated{})
	board.grid[6]["F"] = tile.New(6, "F", tile.Unincorporated{})

	found, corporationTiles = board.TileFoundCorporation(tile.New(6, "E", tile.Unincorporated{}))
	if found {
		t.Errorf("Position %d%s must not found a corporation, got %v instead", 6, "E", corporationTiles)
	}
}

// Testing quadruple merge as this:
//   2 3 4 5 6 7 8 9 1011
// B         []
// C         []
// D         []
// E [][][][]><[][][][][]
// F         []
// G         []
func TestTileQuadrupleMerge(t *testing.T) {
	board := New()
	corp1 := corporation.NewStub("Test 1", 0)
	corp2 := corporation.NewStub("Test 2", 1)
	corp3 := corporation.NewStub("Test 3", 2)
	corp4 := corporation.NewStub("Test 3", 2)
	corp1.SetSize(4)
	corp2.SetSize(5)
	corp3.SetSize(3)
	corp4.SetSize(2)

	board.grid[2]["E"] = tile.New(2, "E", corp1)
	board.grid[3]["E"] = tile.New(3, "E", corp1)
	board.grid[4]["E"] = tile.New(4, "E", corp1)
	board.grid[5]["E"] = tile.New(5, "E", corp1)
	board.grid[7]["E"] = tile.New(7, "E", corp2)
	board.grid[8]["E"] = tile.New(8, "E", corp2)
	board.grid[9]["E"] = tile.New(9, "E", corp2)
	board.grid[10]["E"] = tile.New(10, "E", corp2)
	board.grid[11]["E"] = tile.New(11, "E", corp2)
	board.grid[6]["B"] = tile.New(6, "B", corp3)
	board.grid[6]["C"] = tile.New(6, "C", corp3)
	board.grid[6]["D"] = tile.New(6, "D", corp3)
	board.grid[6]["F"] = tile.New(6, "F", corp4)
	board.grid[6]["G"] = tile.New(6, "G", corp4)

	expectedCorporations := map[string][]corporation.Interface{
		"acquirer": []corporation.Interface{corp2},
		"defunct":  []corporation.Interface{corp1, corp3, corp4},
	}
	merge, corporations := board.TileMergeCorporations(tile.New(6, "E", tile.Unincorporated{}))

	if !slicesSameCorporations(corporations["acquirer"], expectedCorporations["acquirer"]) ||
		!slicesSameCorporations(corporations["defunct"], expectedCorporations["defunct"]) {
		t.Errorf("Position %d%s must merge corporations %v, got %v instead", 6, "E", expectedCorporations, corporations)
	}
	if !merge {
		t.Errorf("TileMergeCorporations() must return true")
	}
}

// Testing quadruple merge tie as this:
//   4 5 6 7 8
// C     []
// D     []
// E [][]><[][]
// F     []
// G     []
func TestTileQuadrupleMergeTie(t *testing.T) {
	board := New()
	corp1 := corporation.NewStub("Test 1", 0)
	corp2 := corporation.NewStub("Test 2", 1)
	corp3 := corporation.NewStub("Test 3", 2)
	corp4 := corporation.NewStub("Test 3", 2)
	corp1.SetSize(2)
	corp2.SetSize(2)
	corp3.SetSize(2)
	corp4.SetSize(2)

	board.grid[4]["E"] = tile.New(4, "E", corp1)
	board.grid[5]["E"] = tile.New(5, "E", corp1)
	board.grid[7]["E"] = tile.New(7, "E", corp2)
	board.grid[8]["E"] = tile.New(8, "E", corp2)
	board.grid[6]["C"] = tile.New(6, "C", corp3)
	board.grid[6]["D"] = tile.New(6, "D", corp3)
	board.grid[6]["F"] = tile.New(6, "F", corp4)
	board.grid[6]["G"] = tile.New(6, "G", corp4)

	expectedCorporations := map[string][]corporation.Interface{
		"acquirer": []corporation.Interface{corp1, corp2, corp3, corp4},
		"defunct":  []corporation.Interface{},
	}
	merge, corporations := board.TileMergeCorporations(tile.New(6, "E", tile.Unincorporated{}))

	if !slicesSameCorporations(corporations["acquirer"], expectedCorporations["acquirer"]) ||
		!slicesSameCorporations(corporations["defunct"], expectedCorporations["defunct"]) {
		t.Errorf("Position %d%s must merge corporations %v, got %v instead", 6, "E", expectedCorporations, corporations)
	}
	if !merge {
		t.Errorf("TileMergeCorporations() must return true")
	}
}

// Testing not a merge as this:
//   3 4 5 6
// E []><[][]
func TestTileDontMerge(t *testing.T) {
	board := New()
	corp2, _ := corporation.New("Test 2", 1)
	board.grid[3]["E"] = tile.New(3, "E", tile.Unincorporated{})
	board.grid[5]["E"] = tile.New(5, "E", corp2)
	board.grid[6]["E"] = tile.New(6, "E", corp2)

	expectedCorporationsMerged := map[string][]corporation.Interface{}
	merge, corporations := board.TileMergeCorporations(tile.New(4, "E", tile.Unincorporated{}))
	if !reflect.DeepEqual(corporations, expectedCorporationsMerged) {
		t.Errorf("Position %d%s must not merge corporations, got %v instead", 4, "E", corporations)
	}
	if merge {
		t.Errorf("TileMergeCorporations() must return false")
	}
}

// Testing growing corporation as this:
//   5 6 7 8
// D   []
// E []><[][]
// F   []
func TestTileGrowCorporation(t *testing.T) {
	board := New()
	corp2, _ := corporation.New("Test 2", 1)
	board.grid[5]["E"] = tile.New(5, "E", tile.Unincorporated{})
	board.grid[7]["E"] = tile.New(7, "E", corp2)
	board.grid[8]["E"] = tile.New(8, "E", corp2)
	board.grid[6]["D"] = tile.New(6, "D", tile.Unincorporated{})
	board.grid[6]["F"] = tile.New(6, "F", tile.Unincorporated{})
	growerTile := tile.New(6, "E", tile.Unincorporated{})

	expectedTilesToAppend := []tile.Interface{
		board.grid[5]["E"],
		board.grid[6]["D"],
		growerTile,
		board.grid[6]["F"],
	}
	expectedCorporationToGrow := corp2
	grow, tilesToAppend, corporationToGrow := board.TileGrowCorporation(growerTile)
	if !slicesSameCells(tilesToAppend, expectedTilesToAppend) {
		t.Errorf(
			"Position %d%s must grow corporation %s by %v, got %v in corporation %s instead",
			6,
			"E",
			expectedCorporationToGrow.Name(),
			expectedTilesToAppend,
			tilesToAppend,
			corporationToGrow.Name(),
		)
	}
	if !grow {
		t.Errorf("TileGrowCorporation() must return true")
	}
}

func TestTileDontGrowCorporation(t *testing.T) {
	board := New()
	corp2, _ := corporation.New("Test 2", 1)

	board.grid[7]["E"] = tile.New(7, "E", corp2)
	board.grid[8]["E"] = tile.New(8, "E", corp2)

	grow, _, _ := board.TileGrowCorporation(tile.New(6, "C", tile.Unincorporated{}))
	if grow {
		t.Errorf(
			"Position %d%s must not grow any corporation, but got true",
			6,
			"C",
		)
	}
}

func TestAdjacentCells(t *testing.T) {
	brd := New()
	tl := tile.New(1, "A", tile.Unincorporated{})
	expectedAdjacentCells := []tile.Interface{
		tile.New(2, "A", tile.Unincorporated{}),
		tile.New(1, "B", tile.Unincorporated{}),
	}

	adjacentCells := brd.AdjacentCells(tl)
	if !slicesSameCells(adjacentCells, expectedAdjacentCells) {
		t.Errorf(
			"Position %d%s expected to have adjacent tiles %v, got %v",
			tl.Number, tl.Letter, expectedAdjacentCells, adjacentCells,
		)
	}
}

func TestSetOwner(t *testing.T) {
	brd := New()
	corp, _ := corporation.New("Test", 1)
	tl1 := tile.New(1, "A", corp)
	tl2 := tile.New(1, "B", corp)
	tls := []tile.Interface{tl1, tl2}
	brd.SetOwner(corp, tls)
	if brd.Cell(tl1.Number(), tl1.Letter()).Owner() != corp || brd.Cell(tl2.Number(), tl2.Letter()).Owner() != corp {
		t.Errorf(
			"Cells %d%s and %d%s expected to belong to corporation",
			tl1.Number(), tl1.Letter(), tl2.Number(), tl2.Letter(),
		)
	}
}

// Compare coordinates of tiles from two slices, order independent
func slicesSameCells(slice1 []tile.Interface, slice2 []tile.Interface) bool {
	if len(slice1) != len(slice2) {
		return false
	}
	var inSlice bool
	for _, val1 := range slice1 {
		inSlice = false
		for _, val2 := range slice2 {
			if val1.Number() == val2.Number() && val1.Letter() == val2.Letter() {
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

// Compare corporations from two slices, order independent
func slicesSameCorporations(slice1 []corporation.Interface, slice2 []corporation.Interface) bool {
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

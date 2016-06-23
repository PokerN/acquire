package fsm

import (
	"github.com/svera/acquire/interfaces"
)

// buyStock is a struct representing a finite state machine's state
type buyStock struct {
	errorState
}

// Name returns state's name
func (s *buyStock) Name() string {
	return interfaces.BuyStockStateName
}

// ToPlayTile returns a PlayTile instance because it's an allowed state transition
func (s *buyStock) ToPlayTile() interfaces.State {
	return &playTile{}
}

// ToEndGame returns an EndGame instance because it's an allowed state transition
func (s *buyStock) ToEndGame() interfaces.State {
	return &endGame{}
}

// ToInsufficientPlayers returns an InsufficientPlayers instance because it's an allowed state transition
func (s *buyStock) ToInsufficientPlayers() interfaces.State {
	return &insufficientPlayers{}
}

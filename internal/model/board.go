package model

import (
	"math/rand"
)

type Ship struct {
	Cells []Position
	Hits  int
}

type Position struct{ X, Y int }

func (s *Ship) IsSunk() bool { return s.Hits >= len(s.Cells) }

type Cell struct {
	Ship *Ship
	Hit  bool
}

type Board struct {
	Grid  [BoardSize][BoardSize]Cell
	Ships []*Ship
}

var fleet = []int{3, 2, 2, 1, 1, 1}

func RandomBoard(rng *rand.Rand) *Board {
	b := &Board{}
	sizes := append([]int(nil), fleet...)
	for _, size := range sizes {
		for {
			horizontal := rng.Intn(2) == 0
			var x, y int
			if horizontal {
				x = rng.Intn(BoardSize - size + 1)
				y = rng.Intn(BoardSize)
			} else {
				x = rng.Intn(BoardSize)
				y = rng.Intn(BoardSize - size + 1)
			}
			if !b.fits(x, y, size, horizontal) {
				continue
			}
			ship := &Ship{}
			for i := 0; i < size; i++ {
				px, py := x, y
				if horizontal {
					px += i
				} else {
					py += i
				}
				ship.Cells = append(ship.Cells, Position{px, py})
				b.Grid[py][px].Ship = ship
			}
			b.Ships = append(b.Ships, ship)
			break
		}
	}
	return b
}

func (b *Board) fits(x, y, size int, horizontal bool) bool {
	for i := 0; i < size; i++ {
		px, py := x, y
		if horizontal {
			px += i
		} else {
			py += i
		}
		if b.Grid[py][px].Ship != nil {
			return false
		}
	}
	return true
}

func (b *Board) Receive(x, y int) (hit, sunk, gameOver bool) {
	cell := &b.Grid[y][x]
	if cell.Hit {
		return cell.Ship != nil, cell.Ship != nil && cell.Ship.IsSunk(), b.allSunk()
	}
	cell.Hit = true
	if cell.Ship == nil {
		return false, false, false
	}
	cell.Ship.Hits++
	return true, cell.Ship.IsSunk(), b.allSunk()
}

func (b *Board) allSunk() bool {
	for _, s := range b.Ships {
		if !s.IsSunk() {
			return false
		}
	}
	return true
}

type ShotResult int

const (
	ShotUnknown ShotResult = iota
	ShotMiss
	ShotHit
)

type ShadowBoard struct {
	Shots [BoardSize][BoardSize]ShotResult
}

func (s *ShadowBoard) Record(x, y int, hit bool) {
	if hit {
		s.Shots[y][x] = ShotHit
	} else {
		s.Shots[y][x] = ShotMiss
	}
}

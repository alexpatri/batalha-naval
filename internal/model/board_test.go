package model

import (
	"math/rand"
	"testing"
)

func TestRandomBoardFleet(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	b := RandomBoard(rng)
	if len(b.Ships) != len(fleet) {
		t.Fatalf("frota %d; quer %d", len(b.Ships), len(fleet))
	}
	sizes := map[int]int{}
	for _, s := range b.Ships {
		sizes[len(s.Cells)]++
	}
	want := map[int]int{3: 1, 2: 2, 1: 3}
	for size, count := range want {
		if sizes[size] != count {
			t.Errorf("tamanho %d: %d navios; quer %d", size, sizes[size], count)
		}
	}
}

func TestRandomBoardNoOverlap(t *testing.T) {
	for seed := int64(0); seed < 50; seed++ {
		rng := rand.New(rand.NewSource(seed))
		b := RandomBoard(rng)
		seen := map[Position]bool{}
		for _, s := range b.Ships {
			for _, p := range s.Cells {
				if seen[p] {
					t.Fatalf("seed %d: posicao %v duplicada", seed, p)
				}
				seen[p] = true
			}
		}
	}
}

func TestReceiveAndGameOver(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	b := RandomBoard(rng)
	hits := 0
	for _, s := range b.Ships {
		for i, p := range s.Cells {
			hit, sunk, gameOver := b.Receive(p.X, p.Y)
			if !hit {
				t.Fatalf("esperava hit em %v", p)
			}
			expectedSunk := i == len(s.Cells)-1
			if sunk != expectedSunk {
				t.Errorf("sunk=%v; quer %v em %v", sunk, expectedSunk, p)
			}
			hits++
			totalCells := 0
			for _, ss := range b.Ships {
				totalCells += len(ss.Cells)
			}
			if gameOver != (hits == totalCells) {
				t.Errorf("gameOver=%v; hits=%d total=%d", gameOver, hits, totalCells)
			}
		}
	}
}

func TestReceiveMiss(t *testing.T) {
	b := &Board{}
	hit, sunk, over := b.Receive(0, 0)
	if hit || sunk || over {
		t.Errorf("miss em board vazio: got hit=%v sunk=%v over=%v", hit, sunk, over)
	}
}

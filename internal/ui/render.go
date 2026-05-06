package ui

import (
	"fmt"
	"strings"

	"batalha-naval/internal/model"
)

const (
	clearScreen = "\033[2J\033[H"
	header      = "   A B C D E F G H I J"
)

func Clear() { fmt.Print(clearScreen) }

func Render(my *model.Board, opp *model.ShadowBoard, status string) {
	Clear()
	var b strings.Builder
	b.WriteString(header)
	b.WriteString("        ")
	b.WriteString(header)
	b.WriteString("\n")
	for row := 0; row < model.BoardSize; row++ {
		fmt.Fprintf(&b, "%2d ", row+1)
		for col := 0; col < model.BoardSize; col++ {
			b.WriteString(myCell(my, col, row))
			b.WriteByte(' ')
		}
		b.WriteString("    ")
		fmt.Fprintf(&b, "%2d ", row+1)
		for col := 0; col < model.BoardSize; col++ {
			b.WriteString(oppCell(opp, col, row))
			b.WriteByte(' ')
		}
		b.WriteByte('\n')
	}
	b.WriteString("\n   tabuleiro: voce")
	b.WriteString(strings.Repeat(" ", 16))
	b.WriteString("tabuleiro: oponente\n\n")
	b.WriteString(status)
	b.WriteByte('\n')
	fmt.Print(b.String())
}

func myCell(b *model.Board, x, y int) string {
	c := b.Grid[y][x]
	switch {
	case c.Hit && c.Ship != nil:
		return "X"
	case c.Hit:
		return "O"
	case c.Ship != nil:
		return "S"
	default:
		return "."
	}
}

func oppCell(s *model.ShadowBoard, x, y int) string {
	switch s.Shots[y][x] {
	case model.ShotHit:
		return "X"
	case model.ShotMiss:
		return "~"
	default:
		return "."
	}
}

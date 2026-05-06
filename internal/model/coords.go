package model

import (
	"fmt"
	"strconv"
	"strings"
)

const BoardSize = 10

func ParseCoord(s string) (x, y int, err error) {
	s = strings.TrimSpace(strings.ToUpper(s))
	if len(s) < 2 || len(s) > 3 {
		return 0, 0, fmt.Errorf("formato inválido: use ex. B7")
	}
	col := s[0]
	if col < 'A' || col >= 'A'+BoardSize {
		return 0, 0, fmt.Errorf("coluna deve ser A-J")
	}
	row, perr := strconv.Atoi(s[1:])
	if perr != nil {
		return 0, 0, fmt.Errorf("linha inválida: %s", s[1:])
	}
	if row < 1 || row > BoardSize {
		return 0, 0, fmt.Errorf("linha deve ser 1-10")
	}
	return int(col - 'A'), row - 1, nil
}

func FormatCoord(x, y int) string {
	return fmt.Sprintf("%c%d", 'A'+x, y+1)
}

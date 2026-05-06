package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"batalha-naval/internal/game"
	"batalha-naval/internal/transport"
)

func main() {
	listen := flag.String("listen", ":9000", "endereco UDP local (ex: :9000)")
	flag.Parse()

	t, err := transport.Listen(*listen)
	if err != nil {
		fmt.Fprintf(os.Stderr, "erro abrindo socket: %v\n", err)
		os.Exit(1)
	}
	defer t.Close()
	fmt.Printf("escutando em %s\n", t.LocalAddr())

	sc := bufio.NewScanner(os.Stdin)
	mode, opponent, err := chooseMode(sc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	e := game.New(t, sc)
	if err := e.Run(ctx, mode, opponent); err != nil {
		fmt.Fprintf(os.Stderr, "fim: %v\n", err)
		os.Exit(1)
	}
}

func chooseMode(sc *bufio.Scanner) (game.Mode, string, error) {
	fmt.Print("(C)onvidar ou (E)sperar? ")
	if !sc.Scan() {
		return 0, "", fmt.Errorf("entrada encerrada")
	}
	switch strings.ToUpper(strings.TrimSpace(sc.Text())) {
	case "C":
		fmt.Print("IP:porta do oponente: ")
		if !sc.Scan() {
			return 0, "", fmt.Errorf("entrada encerrada")
		}
		return game.ModeInvite, strings.TrimSpace(sc.Text()), nil
	case "E":
		return game.ModeWait, "", nil
	default:
		return 0, "", fmt.Errorf("opcao invalida")
	}
}

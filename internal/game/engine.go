package game

import (
	"bufio"
	"context"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"

	"batalha-naval/internal/model"
	"batalha-naval/internal/protocol"
	"batalha-naval/internal/transport"
	"batalha-naval/internal/ui"
)

type Mode int

const (
	ModeInvite Mode = iota
	ModeWait
)

type state int

const (
	stOppTurn state = iota
	stMyTurn
	stOver
)

type Engine struct {
	t       *transport.Transport
	rng     *rand.Rand
	myBoard *model.Board
	shadow  *model.ShadowBoard
	inputCh chan string
	state   state
	won     bool
	status  string
}

func New(t *transport.Transport, sc *bufio.Scanner) *Engine {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	e := &Engine{
		t:       t,
		rng:     rng,
		myBoard: model.RandomBoard(rng),
		shadow:  &model.ShadowBoard{},
		inputCh: make(chan string, 1),
	}
	go func() {
		for sc.Scan() {
			e.inputCh <- strings.TrimSpace(sc.Text())
		}
		close(e.inputCh)
	}()
	return e
}

func (e *Engine) Run(ctx context.Context, mode Mode, opponent string) error {
	switch mode {
	case ModeInvite:
		if err := e.invite(ctx, opponent); err != nil {
			return err
		}
		e.state = stMyTurn
	case ModeWait:
		if err := e.waitForInvite(ctx); err != nil {
			return err
		}
		e.state = stOppTurn
	}

	e.status = initialStatus(e.state)
	ui.Render(e.myBoard, e.shadow, e.status)

	for e.state != stOver {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case in := <-e.t.Recv():
			if err := e.handleIncoming(ctx, in); err != nil {
				return err
			}
		case line := <-e.inputCh:
			if e.state == stMyTurn {
				e.handleAttackInput(ctx, line)
			}
		}
	}
	return nil
}

func initialStatus(s state) string {
	if s == stMyTurn {
		return "sua vez. digite alvo (ex: B7) e tecle enter:"
	}
	return "aguardando jogada do oponente..."
}

func (e *Engine) invite(ctx context.Context, opponent string) error {
	addr, err := net.ResolveUDPAddr("udp", opponent)
	if err != nil {
		return fmt.Errorf("endereco invalido: %w", err)
	}
	e.t.SetPeer(addr)
	fmt.Printf("enviando convite para %s (reenvia a cada 1s ate aceitar)...\n", addr)
	id := e.t.NextID()
	resp, err := e.t.Request(ctx, protocol.WrapInvite(id))
	if err != nil {
		return fmt.Errorf("convite falhou: %w", err)
	}
	ir, ok := resp.Msg.(protocol.InviteResponse)
	if !ok {
		return fmt.Errorf("resposta inesperada: %s", resp.Type)
	}
	if !ir.Accepted {
		return fmt.Errorf("oponente recusou o convite")
	}
	return nil
}

func (e *Engine) waitForInvite(ctx context.Context) error {
	fmt.Printf("aguardando convite em %s...\n", e.t.LocalAddr())
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case in := <-e.t.Recv():
			if in.Env.Type != protocol.TypeInvite {
				continue
			}
			fmt.Printf("convite recebido de %s. aceitar? (s/n): ", in.From)
			line := <-e.inputCh
			accepted := strings.EqualFold(strings.TrimSpace(line), "s")
			if err := in.Reply(protocol.WrapInviteResponse(in.Env.ID, accepted)); err != nil {
				return err
			}
			if !accepted {
				return fmt.Errorf("convite recusado")
			}
			e.t.SetPeer(in.From)
			return nil
		}
	}
}

func (e *Engine) handleIncoming(ctx context.Context, in transport.Incoming) error {
	switch m := in.Env.Msg.(type) {
	case protocol.Attack:
		hit, sunk, gameOver := e.myBoard.Receive(m.Position.X, m.Position.Y)
		if err := in.Reply(protocol.WrapAttackResponse(in.Env.ID, hit, sunk, gameOver)); err != nil {
			return err
		}
		coord := model.FormatCoord(m.Position.X, m.Position.Y)
		switch {
		case gameOver:
			e.status = fmt.Sprintf("oponente atirou em %s e afundou seu ultimo navio. voce perdeu.", coord)
			e.state = stOver
			e.won = false
		case hit:
			extra := ""
			if sunk {
				extra = " (afundou um navio)"
			}
			e.status = fmt.Sprintf("oponente acertou em %s%s. ele atira de novo...", coord, extra)
		default:
			e.status = fmt.Sprintf("oponente errou em %s. sua vez. digite alvo (ex: B7):", coord)
			e.state = stMyTurn
		}
		ui.Render(e.myBoard, e.shadow, e.status)
	default:
	}
	return nil
}

func (e *Engine) handleAttackInput(ctx context.Context, line string) {
	x, y, err := model.ParseCoord(line)
	if err != nil {
		e.status = fmt.Sprintf("entrada invalida (%v). tente de novo (ex: B7):", err)
		ui.Render(e.myBoard, e.shadow, e.status)
		return
	}
	if e.shadow.Shots[y][x] != model.ShotUnknown {
		e.status = "voce ja atirou nessa posicao. escolha outra:"
		ui.Render(e.myBoard, e.shadow, e.status)
		return
	}
	id := e.t.NextID()
	resp, err := e.t.Request(ctx, protocol.WrapAttack(id, x, y))
	if err != nil {
		e.status = fmt.Sprintf("falha ao enviar tiro: %v", err)
		ui.Render(e.myBoard, e.shadow, e.status)
		return
	}
	ar, ok := resp.Msg.(protocol.AttackResponse)
	if !ok {
		e.status = fmt.Sprintf("resposta inesperada: %s", resp.Type)
		ui.Render(e.myBoard, e.shadow, e.status)
		return
	}
	e.shadow.Record(x, y, ar.Hit)
	coord := model.FormatCoord(x, y)
	switch {
	case ar.GameOver:
		e.status = fmt.Sprintf("voce acertou em %s e afundou o ultimo navio. voce venceu!", coord)
		e.state = stOver
		e.won = true
	case ar.Hit:
		extra := ""
		if ar.Sunk {
			extra = " (afundou um navio)"
		}
		e.status = fmt.Sprintf("acertou em %s%s. atire de novo (ex: B7):", coord, extra)
	default:
		e.status = fmt.Sprintf("errou em %s. aguardando jogada do oponente...", coord)
		e.state = stOppTurn
	}
	ui.Render(e.myBoard, e.shadow, e.status)
}

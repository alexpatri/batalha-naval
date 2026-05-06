package transport

import (
	"context"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"batalha-naval/internal/protocol"
)

const (
	retryInterval = 1 * time.Second
	readBuffer    = 64 * 1024
)

type Incoming struct {
	Env   protocol.Envelope
	From  *net.UDPAddr
	Reply func(protocol.Envelope) error
}

type Transport struct {
	conn   *net.UDPConn
	peer   atomic.Pointer[net.UDPAddr]
	nextID atomic.Int64

	recvCh chan Incoming

	mu      sync.Mutex
	pending map[int64]chan protocol.Envelope
	respond map[string]map[int64]*protocol.Envelope
}

func Listen(addr string) (*Transport, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, err
	}
	t := &Transport{
		conn:    conn,
		recvCh:  make(chan Incoming, 16),
		pending: make(map[int64]chan protocol.Envelope),
		respond: make(map[string]map[int64]*protocol.Envelope),
	}
	t.nextID.Store(0)
	go t.readLoop()
	return t, nil
}

func (t *Transport) Close() error      { return t.conn.Close() }
func (t *Transport) LocalAddr() string { return t.conn.LocalAddr().String() }

func (t *Transport) SetPeer(p *net.UDPAddr) { t.peer.Store(p) }
func (t *Transport) Peer() *net.UDPAddr     { return t.peer.Load() }

func (t *Transport) NextID() int64 { return t.nextID.Add(1) }

func (t *Transport) Recv() <-chan Incoming { return t.recvCh }

func (t *Transport) sendTo(addr *net.UDPAddr, env protocol.Envelope) error {
	data, err := protocol.Encode(env)
	if err != nil {
		return err
	}
	_, err = t.conn.WriteToUDP(data, addr)
	return err
}

func (t *Transport) Send(env protocol.Envelope) error {
	peer := t.Peer()
	if peer == nil {
		return fmt.Errorf("peer not set")
	}
	return t.sendTo(peer, env)
}

func (t *Transport) Request(ctx context.Context, env protocol.Envelope) (protocol.Envelope, error) {
	peer := t.Peer()
	if peer == nil {
		return protocol.Envelope{}, fmt.Errorf("peer not set")
	}
	return t.requestTo(ctx, peer, env)
}

func (t *Transport) RequestTo(ctx context.Context, addr *net.UDPAddr, env protocol.Envelope) (protocol.Envelope, error) {
	return t.requestTo(ctx, addr, env)
}

func (t *Transport) requestTo(ctx context.Context, addr *net.UDPAddr, env protocol.Envelope) (protocol.Envelope, error) {
	ch := make(chan protocol.Envelope, 1)
	t.mu.Lock()
	t.pending[env.ID] = ch
	t.mu.Unlock()
	defer func() {
		t.mu.Lock()
		delete(t.pending, env.ID)
		t.mu.Unlock()
	}()

	if err := t.sendTo(addr, env); err != nil {
		return protocol.Envelope{}, err
	}
	ticker := time.NewTicker(retryInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return protocol.Envelope{}, ctx.Err()
		case resp := <-ch:
			return resp, nil
		case <-ticker.C:
			if err := t.sendTo(addr, env); err != nil {
				return protocol.Envelope{}, err
			}
		}
	}
}

func (t *Transport) readLoop() {
	buf := make([]byte, readBuffer)
	for {
		n, src, err := t.conn.ReadFromUDP(buf)
		if err != nil {
			return
		}
		env, err := protocol.Decode(buf[:n])
		if err != nil {
			continue
		}
		t.dispatch(env, src)
	}
}

func (t *Transport) dispatch(env protocol.Envelope, src *net.UDPAddr) {
	t.mu.Lock()
	if ch, ok := t.pending[env.ID]; ok {
		delete(t.pending, env.ID)
		t.mu.Unlock()
		select {
		case ch <- env:
		default:
		}
		return
	}
	key := src.String()
	if entry, seen := t.respond[key][env.ID]; seen {
		t.mu.Unlock()
		if entry != nil {
			_ = t.sendTo(src, *entry)
		}
		return
	}
	if t.respond[key] == nil {
		t.respond[key] = make(map[int64]*protocol.Envelope)
	}
	t.respond[key][env.ID] = nil
	t.mu.Unlock()

	reply := func(resp protocol.Envelope) error {
		t.mu.Lock()
		if t.respond[key] == nil {
			t.respond[key] = make(map[int64]*protocol.Envelope)
		}
		cached := resp
		t.respond[key][env.ID] = &cached
		t.mu.Unlock()
		return t.sendTo(src, resp)
	}
	t.recvCh <- Incoming{Env: env, From: src, Reply: reply}
}

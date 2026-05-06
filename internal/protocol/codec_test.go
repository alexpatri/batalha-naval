package protocol

import "testing"

func TestRoundTripInvite(t *testing.T) {
	env := WrapInvite(42)
	data, err := Encode(env)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	if got.Type != TypeInvite || got.ID != 42 {
		t.Errorf("got %+v", got)
	}
}

func TestRoundTripAttackResponse(t *testing.T) {
	env := WrapAttackResponse(7, true, true, false)
	data, err := Encode(env)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	ar, ok := got.Msg.(AttackResponse)
	if !ok {
		t.Fatalf("tipo errado: %T", got.Msg)
	}
	if ar.ID != 7 || !ar.Hit || !ar.Sunk || ar.GameOver {
		t.Errorf("got %+v", ar)
	}
}

func TestDecodeUnknownType(t *testing.T) {
	_, err := Decode([]byte(`{"id":1,"type":"FOO"}`))
	if err == nil {
		t.Error("esperava erro para tipo desconhecido")
	}
}

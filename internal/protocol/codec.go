package protocol

import (
	"encoding/json"
	"fmt"
)

func Encode(env Envelope) ([]byte, error) {
	return json.Marshal(env.Msg)
}

func Decode(data []byte) (Envelope, error) {
	var head struct {
		ID   int64  `json:"id"`
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &head); err != nil {
		return Envelope{}, fmt.Errorf("decode head: %w", err)
	}
	env := Envelope{ID: head.ID, Type: head.Type}
	switch head.Type {
	case TypeInvite:
		var m Invite
		if err := json.Unmarshal(data, &m); err != nil {
			return Envelope{}, err
		}
		env.Msg = m
	case TypeInviteResponse:
		var m InviteResponse
		if err := json.Unmarshal(data, &m); err != nil {
			return Envelope{}, err
		}
		env.Msg = m
	case TypeAttack:
		var m Attack
		if err := json.Unmarshal(data, &m); err != nil {
			return Envelope{}, err
		}
		env.Msg = m
	case TypeAttackResponse:
		var m AttackResponse
		if err := json.Unmarshal(data, &m); err != nil {
			return Envelope{}, err
		}
		env.Msg = m
	default:
		return Envelope{}, fmt.Errorf("unknown type %q", head.Type)
	}
	return env, nil
}

func WrapInvite(id int64) Envelope {
	m := Invite{ID: id, Type: TypeInvite}
	return Envelope{ID: id, Type: TypeInvite, Msg: m}
}

func WrapInviteResponse(id int64, accepted bool) Envelope {
	m := InviteResponse{ID: id, Type: TypeInviteResponse, Accepted: accepted}
	return Envelope{ID: id, Type: TypeInviteResponse, Msg: m}
}

func WrapAttack(id int64, x, y int) Envelope {
	m := Attack{ID: id, Type: TypeAttack, Position: Position{X: x, Y: y}}
	return Envelope{ID: id, Type: TypeAttack, Msg: m}
}

func WrapAttackResponse(id int64, hit, sunk, gameOver bool) Envelope {
	m := AttackResponse{
		ID:       id,
		Type:     TypeAttackResponse,
		Hit:      hit,
		Sunk:     sunk,
		GameOver: gameOver,
	}
	return Envelope{ID: id, Type: TypeAttackResponse, Msg: m}
}

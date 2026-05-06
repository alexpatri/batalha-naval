package protocol

const (
	TypeInvite          = "INVITE"
	TypeInviteResponse  = "INVITE_RESPONSE"
	TypeAttack          = "ATTACK"
	TypeAttackResponse  = "ATTACK_RESPONSE"
)

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Invite struct {
	ID   int64  `json:"id"`
	Type string `json:"type"`
}

type InviteResponse struct {
	ID       int64  `json:"id"`
	Type     string `json:"type"`
	Accepted bool   `json:"accepted"`
}

type Attack struct {
	ID       int64    `json:"id"`
	Type     string   `json:"type"`
	Position Position `json:"position"`
}

type AttackResponse struct {
	ID       int64  `json:"id"`
	Type     string `json:"type"`
	Hit      bool   `json:"hit"`
	Sunk     bool   `json:"sunk"`
	GameOver bool   `json:"gameOver"`
}

type Envelope struct {
	ID   int64
	Type string
	Msg  any
}

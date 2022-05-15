package app

type CardEffectContext struct {
	Card    *Card
	Game    *Game
	Monster *MonsterCard
	Effect  *CardEffect
	Belong  *Player
}

func (m *CardEffectContext) IsCurrentTurn() bool {
	return m.Game.TurnCount == m.Effect.CreateTurnCount
}

func (m *CardEffectContext) IsSelfTurn() bool {
	return m.Game.CurrentPlayer == m.Belong
}

func (m *CardEffectContext) GetOpponentEvoSourceMonsterCount() int {
	area := m.Game.PlayerAreas[m.Belong.Opponent.Session.ID].Field
	count := 0
	for _, monster := range area {
		if len(monster.List) > 1 {
			count++
		}
	}
	return count
}
func (m *CardEffectContext) GetOpponentNoEvoSourceMonsterCount() int {
	area := m.Game.PlayerAreas[m.Belong.Opponent.Session.ID].Field
	return len(area) - m.GetOpponentEvoSourceMonsterCount()
}
func (m *CardEffectContext) GetSelfDefenseCount() int {
	area := m.Game.PlayerAreas[m.Belong.Opponent.Session.ID].Defense
	return len(area)
}
func (m *CardEffectContext) GetOpponentSleepCount() int {
	area := m.Game.PlayerAreas[m.Belong.Opponent.Session.ID].Field
	count := 0
	for _, monster := range area {
		if monster.Sleep {
			count++
		}
	}
	return count
}

type CardEffect struct {
	IsEvoSource     bool                         //是否为进化源
	ActiveTime      string                       //触发时点 attack, defend, draw, discard, play, evolve
	Action          func(ctx *CardEffectContext) `json:"-"`
	CreateTurnCount int                          //创建该效果的回合数
}

type CardEffectManager struct {
	Effects map[string][]CardEffect
}

func (m *CardEffectManager) RegistEffect(serial string, e CardEffect) {
	if _, ok := m.Effects[serial]; !ok {
		m.Effects[serial] = []CardEffect{}
	} else {
		m.Effects[serial] = append(m.Effects[serial], e)
	}
}

func NewCardEffectManager() *CardEffectManager {
	m := &CardEffectManager{
		Effects: make(map[string][]CardEffect),
	}
	m.RegistEffect("BT1-001", CardEffect{
		IsEvoSource: true,
		ActiveTime:  "attack",
		Action: func(ctx *CardEffectContext) {
			ctx.Card.EffectList = append(ctx.Card.EffectList, CardEffect{
				CreateTurnCount: ctx.Game.TurnCount,
				Action: func(ctx2 *CardEffectContext) {
					if !ctx2.IsCurrentTurn() {
						return
					}
					ctx2.Monster.DP += 1000
				},
			})
		},
	})
	m.RegistEffect("BT1-002", CardEffect{
		IsEvoSource: true,
		Action: func(ctx *CardEffectContext) {
			if !ctx.IsSelfTurn() {
				return
			}
			ctx.Monster.DP += 2000
		},
	})

	//TODO 一回合一次
	m.RegistEffect("BT1-003", CardEffect{
		IsEvoSource: true,
		ActiveTime:  "attack",
		Action: func(ctx *CardEffectContext) {
			count := ctx.GetOpponentEvoSourceMonsterCount()
			if count == 0 {
				ctx.Game.Draw(ctx.Belong, 1)
			}
		},
	})
	m.RegistEffect("BT1-004", CardEffect{
		IsEvoSource: true,
		Action: func(ctx *CardEffectContext) {
			if !ctx.IsSelfTurn() {
				return
			}
			count := ctx.GetOpponentNoEvoSourceMonsterCount()
			if count >= 2 {
				ctx.Monster.DP += 2000
			}
		},
	})
	m.RegistEffect("BT1-005", CardEffect{
		IsEvoSource: true,
		Action: func(ctx *CardEffectContext) {
			if !ctx.IsSelfTurn() {
				return
			}
			count := ctx.GetSelfDefenseCount()
			if count >= 6 {
				ctx.Monster.DP += 2000
			}
		},
	})
	m.RegistEffect("BT1-006", CardEffect{
		IsEvoSource: true,
		ActiveTime:  "attack",
		Action: func(ctx *CardEffectContext) {
			count := ctx.GetSelfDefenseCount()
			if count >= 5 {
				ctx.Game.Draw(ctx.Belong, 1)
			}
		},
	})
	m.RegistEffect("BT1-007", CardEffect{
		IsEvoSource: true,
		ActiveTime:  "attack",
		Action: func(ctx *CardEffectContext) {
			if ctx.Game.EachTurnEvoCount == 0 {
				return
			}
			ctx.Card.EffectList = append(ctx.Card.EffectList, CardEffect{
				CreateTurnCount: ctx.Game.TurnCount,
				Action: func(ctx2 *CardEffectContext) {
					if !ctx2.IsCurrentTurn() {
						return
					}
					ctx2.Monster.DP += 1000
				},
			})
		},
	})
	m.RegistEffect("BT1-008", CardEffect{
		IsEvoSource: true,
		Action: func(ctx *CardEffectContext) {
			if !ctx.IsSelfTurn() {
				return
			}
			count := ctx.GetOpponentSleepCount()
			if count >= 2 {
				ctx.Monster.DP += 2000
			}
		},
	})
	m.RegistEffect("BT1-010", CardEffect{
		ActiveTime: "summon",
		Action: func(ctx *CardEffectContext) {
			//todo
		},
	})
	m.RegistEffect("BT1-011", CardEffect{
		ActiveTime: "summon",
		Action: func(ctx *CardEffectContext) {
			//todo
		},
	})
	return m
}

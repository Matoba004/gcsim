package fischl

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func (c *char) c6Wave() {
	ai := info.AttackInfo{
		ActorIndex: c.Index(),
		Abil:       "Evernight Raven (C6)",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagElementalArt,
		ICDGroup:   attacks.ICDGroupFischl,
		StrikeType: attacks.StrikeTypePierce,
		Element:    attributes.Electro,
		Durability: 25,
		Mult:       0.3,
	}

	// C6 uses Oz Snapshot
	c.Core.QueueAttackWithSnap(
		ai,
		c.ozSnapshot.Snapshot,
		combat.NewBoxHit(
			c.Core.Combat.Player(),
			c.Core.Combat.PrimaryTarget(),
			info.Point{Y: -1},
			0.1,
			1,
		),
		c.ozTravel,
	)

	c.c6Hexerei()
}

func (c *char) c6Hexerei() {
	if c.Hexerei != 1 {
		return
	}

	if c.Core.Player.GetHexereiCount() < 2 {
		return
	}
	if c.StatModIsActive("fischl-hex-overload") {
		duration := c.StatusDuration("fischl-hex-overload")
		m := make([]float64, attributes.EndStatType)
		m[attributes.ATKP] = 0.225

		c.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("fischl-hex-overload-c6", duration),
			AffectedStat: attributes.ATKP,
			Amount: func() []float64 {
				return m
			},
		})

		c.Core.Player.ActiveChar().AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("fischl-hex-overload-c6", duration),
			AffectedStat: attributes.EM,
			Amount: func() []float64 {
				return m
			},
		})
	}

	if c.StatModIsActive("fischl-hex-electrocharged") {
		duration := c.StatusDuration("fischl-hex-electrocharged")
		m := make([]float64, attributes.EndStatType)
		m[attributes.EM] = 90

		c.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("fischl-hex-electrocharged-c6", duration),
			AffectedStat: attributes.EM,
			Amount: func() []float64 {
				return m
			},
		})

		c.Core.Player.ActiveChar().AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("fischl-hex-electrocharged-c6", duration),
			AffectedStat: attributes.EM,
			Amount: func() []float64 {
				return m
			},
		})
	}
}

package fischl

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const a4IcdKey = "fischl-a4-icd"

// A1 is not implemented:
// TODO: When Fischl hits Oz with a fully-charged Aimed Shot, Oz brings down Thundering Retribution, dealing AoE Electro DMG equal to 152.7% of the arrow's DMG.

// If your current active character triggers an Electro-related Elemental Reaction when Oz is on the field,
// the opponent shall be stricken with Thundering Retribution that deals Electro DMG equal to 80% of Fischl's ATK.
func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}

	// Hyperbloom comes from a gadget so it doesn't ignore gadgets
	a4cb := func(args ...any) {
		ae := args[1].(*info.AttackEvent)

		if ae.Info.ActorIndex != c.Core.Player.Active() {
			return
		}
		// do nothing if oz not on field
		if !c.StatusIsActive(ozActiveKey) {
			return
		}
		active := c.Core.Player.ActiveChar()
		if active.StatusIsActive(a4IcdKey) {
			return
		}
		active.AddStatus(a4IcdKey, 0.5*60, true)

		ai := info.AttackInfo{
			ActorIndex: c.Index(),
			Abil:       "Thundering Retribution (A4)",
			AttackTag:  attacks.AttackTagElementalArt,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupFischl,
			StrikeType: attacks.StrikeTypePierce,
			Element:    attributes.Electro,
			Durability: 25,
			Mult:       0.8,
		}

		// A4 uses Oz Snapshot
		// TODO: this should target closest enemy within 15m of "elemental reaction position"
		c.Core.QueueAttackWithSnap(
			ai,
			c.ozSnapshot.Snapshot,
			combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 0.5),
			4)
	}
	a4cbNoGadget := func(args ...any) {
		if _, ok := args[0].(*enemy.Enemy); ok {
			a4cb(args...)
		}
	}

	c.Core.Events.Subscribe(event.OnOverload, a4cbNoGadget, "fischl-a4")
	c.Core.Events.Subscribe(event.OnElectroCharged, a4cbNoGadget, "fischl-a4")
	c.Core.Events.Subscribe(event.OnLunarCharged, a4cbNoGadget, "fischl-a4")
	c.Core.Events.Subscribe(event.OnSuperconduct, a4cbNoGadget, "fischl-a4")
	c.Core.Events.Subscribe(event.OnSwirlElectro, a4cbNoGadget, "fischl-a4")
	c.Core.Events.Subscribe(event.OnCrystallizeElectro, a4cbNoGadget, "fischl-a4")
	c.Core.Events.Subscribe(event.OnHyperbloom, a4cb, "fischl-a4")
	c.Core.Events.Subscribe(event.OnQuicken, a4cbNoGadget, "fischl-a4")
	c.Core.Events.Subscribe(event.OnAggravate, a4cbNoGadget, "fischl-a4")
}

func (c *char) witchesEveRite() {
	// if is hexerei
	if c.Hexerei != 1 {
		return
	}

	if c.Core.Player.GetHexereiCount() < 2 {
		return
	}

	olF := func(args ...any) {
		m := make([]float64, attributes.EndStatType)
		m[attributes.ATKP] = 0.225

		c.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("fischl-hex-overload", 10*60),
			AffectedStat: attributes.ATKP,
			Amount: func() []float64 {
				return m
			},
		})

		c.Core.Player.ActiveChar().AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("fischl-hex-overload", 10*60),
			AffectedStat: attributes.ATKP,
			Amount: func() []float64 {
				return m
			},
		})
	}

	ecF := func(args ...any) {
		m := make([]float64, attributes.EndStatType)
		m[attributes.EM] = 90

		c.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("fischl-hex-electrocharged", 10*60),
			AffectedStat: attributes.EM,
			Amount: func() []float64 {
				return m
			},
		})

		c.Core.Player.ActiveChar().AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("fischl-hex-electrocharged", 10*60),
			AffectedStat: attributes.EM,
			Amount: func() []float64 {
				return m
			},
		})
	}

	updateBuffForActive := func(args ...any) {
		// swap should update buff for new active char and remove from old active char
		prev := args[0].(int)
		next := args[1].(int)

		statMods := []string{"fischl-hex-overload", "fischl-hex-electrocharged", "fischl-hex-overload-c6", "fischl-hex-electrocharged-c6"}

		for _, mod := range statMods {
			if c.StatModIsActive(mod) {
				if prev != c.Index() {
					c.Core.Player.Chars()[prev].DeleteStatMod(mod)
				}
				m := make([]float64, attributes.EndStatType)

				if mod == "fischl-hex-overload" || mod == "fischl-hex-overload-c6" {
					m[attributes.ATKP] = 0.225
				} else {
					m[attributes.EM] = 90
				}

				c.Core.Player.Chars()[next].AddStatMod(character.StatMod{
					Base: modifier.NewBaseWithHitlag(mod, 10*60),
					Amount: func() []float64 {
						return m
					},
				})
			}
		}

		if c.StatModIsActive("fischl-hex-electrocharged") {
			if prev != c.Index() {
				c.Core.Player.Chars()[prev].DeleteStatMod("fischl-hex-electrocharged")
			}
			m := make([]float64, attributes.EndStatType)
			m[attributes.EM] = 90

			c.Core.Player.Chars()[next].AddStatMod(character.StatMod{
				Base:         modifier.NewBaseWithHitlag("fischl-hex-electrocharged", 10*60),
				AffectedStat: attributes.EM,
				Amount: func() []float64 {
					return m
				},
			})
		}
	}

	c.Core.Events.Subscribe(event.OnOverload, olF, "fischl-hex-overload")
	c.Core.Events.Subscribe(event.OnElectroCharged, ecF, "fischl-hex-electrocharged")
	c.Core.Events.Subscribe(event.OnCharacterSwap, updateBuffForActive, "fischl-hex-update-buff")
}

package breezebornerefrain

import (
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const (
	stackGainICDKey = "hymn-of-the-pure-stack-gain-icd"
	stellarBonusKey = "breezeborne-refrain-stellar-bonus"
)

type Weapon struct {
	Index               int
	hymnOfThePureStacks int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{
		hymnOfThePureStacks: 0,
	}

	erBuff := make([]float64, attributes.EndStatType)
	erBuff[attributes.ER] = 0.15 + float64(p.Refine)*0.05

	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBaseWithHitlag("breezeborne-refrain-er", -1),
		AffectedStat: attributes.ER,
		Amount: func() []float64 {
			return erBuff
		},
	})

	stellarBuff := 0.18 + float64(p.Refine)*0.06

	onSkillOrBurstHit := func(args ...any) {
		if char.StatModIsActive(stellarBonusKey) {
			return
		}

		if char.StatModIsActive(stackGainICDKey) {
			return
		}

		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex != char.Index() {
			return
		}

		switch atk.Info.AttackTag {
		case attacks.AttackTagElementalArt:
		case attacks.AttackTagElementalBurst:
		default:
			return
		}

		w.hymnOfThePureStacks++

		char.AddStatus(stackGainICDKey, 2, true) // 0.03s ICD

		if w.hymnOfThePureStacks >= 3 {
			w.hymnOfThePureStacks = 0
			for _, char := range c.Player.Chars() {
				char.AddReactBonusMod(character.ReactBonusMod{
					Base: modifier.NewBaseWithHitlag(stellarBonusKey, 12*60),
					Amount: func(atk info.AttackInfo) float64 {
						switch atk.AttackTag {
						case attacks.AttackTagDirectStellarConduct:
						case attacks.AttackTagDirectStellarSwirl:
						default:
							return 0
						}

						return stellarBuff
					},
				})
			}
		}
	}

	c.Events.Subscribe(event.OnEnemyHit, onSkillOrBurstHit, "breezeborne-refrain-on-skill-or-burst-hit-"+char.Base.Key.String())

	return w, nil
}

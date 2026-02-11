package sucrose

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

var skillFrames []int

const particleICDKey = "sucrose-particle-icd"

const skillHexBuffKey = "sucrose-hex-skill"

func init() {
	skillFrames = frames.InitAbilSlice(68) // walk
	skillFrames[action.ActionAttack] = 57
	skillFrames[action.ActionCharge] = 56
	skillFrames[action.ActionSkill] = 56
	skillFrames[action.ActionBurst] = 57
	skillFrames[action.ActionDash] = 11
	skillFrames[action.ActionJump] = 11
	skillFrames[action.ActionSwap] = 56
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	ai := info.AttackInfo{
		ActorIndex: c.Index(),
		Abil:       "Astable Anemohypostasis Creation-6308",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Anemo,
		Durability: 25,
		Mult:       skill[c.TalentLvlSkill()],
	}

	done := false
	a4CB := func(a info.AttackCB) {
		if a.Target.Type() != info.TargettableEnemy {
			return
		}
		if done {
			return
		}
		done = true
		c.a4()
	}

	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.Player(), info.Point{Y: 5}, 6),
		0,
		42,
		a4CB,
		c.particleCB,
	)

	// reduce charge by 1
	c.SetCDWithDelay(action.ActionSkill, 900, 9)
	c.witchesEveRiteSkill()

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionDash], // earliest cancel
		State:           action.SkillState,
	}, nil
}

func (c *char) particleCB(a info.AttackCB) {
	if a.Target.Type() != info.TargettableEnemy {
		return
	}
	if c.StatusIsActive(particleICDKey) {
		return
	}
	c.AddStatus(particleICDKey, 0.4*60, false)
	c.Core.QueueParticle(c.Base.Key.String(), 4, attributes.Anemo, c.ParticleDelay)
}

func (c *char) witchesEveRiteSkill() {
	if c.Hexerei != 1 {
		return
	}

	if c.Core.Player.GetHexereiCount() < 2 {
		return
	}

	m := make([]float64, attributes.EndStatType)
	m[attributes.DmgP] = 0.4 / 7

	for _, char := range c.Core.Player.Chars() {
		char.AddAttackMod(character.AttackMod{
			Base: modifier.NewBaseWithHitlag(skillHexBuffKey, 60*10),
			Amount: func(atk *info.AttackEvent, t info.Target) []float64 {
				switch atk.Info.AttackTag {
				case attacks.AttackTagNormal:
				case attacks.AttackTagExtra:
				case attacks.AttackTagPlunge:
				case attacks.AttackTagElementalArt:
				case attacks.AttackTagElementalArtHold:
				case attacks.AttackTagElementalBurst:
				default:
					return nil
				}

				return m
			},
		})
	}
}

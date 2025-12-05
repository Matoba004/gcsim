package iansan

import (
	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/internal/template/nightsoul"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
)

func init() {
	core.RegisterCharFunc(keys.Iansan, NewChar)
}

type char struct {
	*tmpl.Character
	nightsoulState *nightsoul.State

	nightsoulSrc      int
	particleGenerated bool
	burstSrc          int
	burstBuff         []float64
	burstRestoreNS    int
	pointsOverflow    float64

	a1Increase bool
	a4Src      int

	c1Points float64
	c4Stacks int

	nightsoulDecrease float64
}

func NewChar(s *core.Core, w *character.CharWrapper, p info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.EnergyMax = 70
	c.SkillCon = 3
	c.BurstCon = 5
	c.NormalHitNum = normalHitNum

	c.nightsoulState = nightsoul.New(s, w)
	c.nightsoulState.MaxPoints = 54

	nb, ok := p.Params["nightsoul_bleed"]

	nightsoulDecrease := float64(nb) / 100.0 * maxNightsoulDecrease

	if !ok {
		nightsoulDecrease = maxNightsoulDecrease
	}
	nightsoulDecrease = max(min(nightsoulDecrease, maxNightsoulDecrease), 0)
	c.nightsoulDecrease = float64(nightsoulDecrease)

	w.Character = &c

	return nil
}

func (c *char) Init() error {
	c.burstBuff = make([]float64, attributes.EndStatType)

	c.a1()
	c.a4()

	return nil
}

func (c *char) ActionStam(a action.Action, p map[string]int) float64 {
	if a == action.ActionCharge && c.StatusIsActive(fastSkill) {
		return 0
	}
	return c.Character.ActionStam(a, p)
}

func (c *char) Condition(fields []string) (any, error) {
	switch fields[0] {
	case "nightsoul":
		return c.nightsoulState.Condition(fields)
	default:
		return c.Character.Condition(fields)
	}
}

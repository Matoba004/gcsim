package reactable_test

import (
	"math"
	"testing"

	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/enemy"
)

func TestNonMutableVape(t *testing.T) {
	c, _ := makeCore(0)

	// create enemy with hydro aura
	trg := enemy.New(c, info.EnemyProfile{
		Level:  100,
		Resist: make(map[attributes.Element]float64),
		Pos: info.Coord{
			X: 0,
			Y: 0,
			R: 1,
		},
		Element:           attributes.Hydro,
		ElementDurability: 25,
	})
	c.Combat.AddEnemy(trg)

	err := c.Init()
	if err != nil {
		t.Errorf("error initializing core: %v", err)
		t.FailNow()
	}

	count := 0

	incrementCounter := func(args ...any) {
		count++
	}

	c.Events.Subscribe(event.OnVaporize, incrementCounter, "vaporize")

	c.QueueAttackEvent(&info.AttackEvent{
		Info: info.AttackInfo{
			Element:    attributes.Pyro,
			Durability: 25,
		},
		Pattern: combat.NewCircleHitOnTarget(info.Point{}, nil, 100),
	}, 0)
	advanceCoreFrame(c)

	if float64(trg.GetDurability()[attributes.Pyro]) > 0.000001 {
		t.Errorf(
			"expected pyro=%v, got pyro=%v",
			0,
			trg.GetDurability()[attributes.Pyro],
		)
	}
	if math.Abs(float64(trg.GetDurability()[attributes.Hydro])-25) > 0.000001 {
		t.Errorf(
			"expected hydro=%v, got hydro=%v",
			25,
			trg.GetDurability()[attributes.Hydro],
		)
	}
	if count != 1 {
		t.Errorf(
			"expected %v vaporizes, got %v",
			1,
			count,
		)
	}
}

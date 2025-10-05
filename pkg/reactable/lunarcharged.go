package reactable

import (
	"fmt"
	"sort"

	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/glog"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

type ContributorData struct {
	damage []float64
}

func (r *Reactable) TryAddLC(a *info.AttackEvent) bool {
	if a.Info.Durability < info.ZeroDur {
		return false
	}
	// if there's still frozen left don't try to ec
	// game actively rejects ec reaction if frozen is present
	if r.Durability[info.ReactionModKeyFrozen] > info.ZeroDur {
		return false
	}

	// adding ec or hydro just adds to durability
	switch a.Info.Element {
	case attributes.Hydro:
		// if there's no existing hydro or electro then do nothing
		if r.Durability[info.ReactionModKeyElectro] < info.ZeroDur {
			return false
		}
		// add to hydro durability (can't add if the atk already reacted)
		// TODO: this shouldn't happen here
		if !a.Reacted {
			r.attachOrRefillNormalEle(info.ReactionModKeyHydro, a.Info.Durability)
		}
	case attributes.Electro:
		// if there's no existing hydro or electro then do nothing
		if r.Durability[info.ReactionModKeyHydro] < info.ZeroDur {
			return false
		}
		// add to electro durability (can't add if the atk already reacted)
		if !a.Reacted {
			r.attachOrRefillNormalEle(info.ReactionModKeyElectro, a.Info.Durability)
		}
	default:
		return false
	}

	a.Reacted = true
	r.core.Events.Emit(event.OnElectroCharged, r.self, a)

	// at this point ec is refereshed so we need to trigger a reaction
	// and change ownership
	atk := info.AttackInfo{
		ActorIndex:       a.Info.ActorIndex,
		DamageSrc:        r.self.Key(),
		Abil:             string(info.ReactionTypeLunarCharged),
		AttackTag:        attacks.AttackTagECDamage,
		ICDTag:           attacks.ICDTagECDamage,
		ICDGroup:         attacks.ICDGroupReactionB,
		StrikeType:       attacks.StrikeTypeDefault,
		Element:          attributes.Electro,
		IgnoreDefPercent: 1,
	}

	// add damage additions from all contributors

	// first contributor does 100% dmg, second 1/2, third and fourth 1/12

	atk.FlatDmg = r.calcFinalLCDmg(atk)

	r.ecAtk = atk

	// if this is a new ec then trigger tick immediately and queue up ticks
	// otherwise do nothing
	// TODO: need to check if refresh ec triggers new tick immediately or not
	if r.ecTickSrc == -1 {
		r.ecTickSrc = r.core.F
		r.core.QueueAttackWithSnap(
			r.ecAtk,
			r.ecSnapshot,
			combat.NewSingleTargetHit(r.self.Key()),
			10,
		)

		r.core.Tasks.Add(r.nextTickLC(r.core.F), 120+10)
		// subscribe to wane ticks
		r.core.Events.Subscribe(event.OnEnemyDamage, func(args ...any) bool {
			// target should be first, then snapshot
			n := args[0].(info.Target)
			a := args[1].(*info.AttackEvent)
			dmg := args[2].(float64)
			// TODO: there's no target index
			if n.Key() != r.self.Key() {
				return false
			}
			if a.Info.AttackTag != attacks.AttackTagLCDamage {
				return false
			}
			// ignore if this dmg instance has been wiped out due to icd
			if dmg == 0 {
				return false
			}
			// ignore if we no longer have both electro and hydro
			if r.Durability[info.ReactionModKeyElectro] < info.ZeroDur || r.Durability[info.ReactionModKeyHydro] < info.ZeroDur {
				return true
			}

			// wane in 0.1 seconds
			r.core.Tasks.Add(func() {
				r.waneLC()
			}, 6)
			return false
		}, fmt.Sprintf("lc-%v", r.self.Key()))
	}

	// ticks are 60 frames since last tick
	// taking tick dmg resets last tick
	return true
}

func (r *Reactable) calcFinalLCDmg(atk info.AttackInfo) float64 {
	contributor_dmg := [4]float64{0.0, 0.0, 0.0, 0.0}
	index := 0
	for frame, v := range r.contributors {
		// if contributor expired skip them
		// contributor expires after 120 frames (2 seconds)
		// frame is when contributor was added + 120
		// so if current frame is > frame then contributor expired
		if r.core.F > frame {
			delete(r.contributors, frame)
			continue
		}
		char := r.core.Player.ByIndex(v.ActorIndex)
		em := char.Stat(attributes.EM)
		flatdmg := combat.CalcLunarChargedDmg(char.Base.Level, char, atk, em) * 1.8
		if r.core.Rand.Float64() <= char.Stat(attributes.CR) {
			flatdmg *= (1 + char.Stat(attributes.CD))
		}
		contributor_dmg[index] = flatdmg
		index++
	}
	// sort contributors by dmg
	sort.Slice(contributor_dmg[:], func(i, j int) bool {
		return contributor_dmg[i] > contributor_dmg[j]
	})

	return contributor_dmg[0] + (1.0/2.0)*contributor_dmg[1] + (1.0/12.0)*contributor_dmg[2] + (1.0/12.0)*contributor_dmg[3]
}

func (r *Reactable) waneLC() {
	r.Durability[info.ReactionModKeyElectro] -= 10
	r.Durability[info.ReactionModKeyElectro] = max(0, r.Durability[info.ReactionModKeyElectro])
	r.Durability[info.ReactionModKeyHydro] -= 10
	r.Durability[info.ReactionModKeyHydro] = max(0, r.Durability[info.ReactionModKeyHydro])
	r.core.Log.NewEvent("lc wane",
		glog.LogElementEvent,
		-1,
	).
		Write("aura", "lc").
		Write("target", r.self.Key()).
		Write("hydro", r.Durability[info.ReactionModKeyHydro]).
		Write("electro", r.Durability[info.ReactionModKeyElectro])

	// lc is gone
	r.checkLC()
}

func (r *Reactable) checkLC() {
	if r.Durability[info.ReactionModKeyElectro] < info.ZeroDur || r.Durability[info.ReactionModKeyHydro] < info.ZeroDur {
		r.ecTickSrc = -1
		r.core.Events.Unsubscribe(event.OnEnemyDamage, fmt.Sprintf("lc-%v", r.self.Key()))
		r.core.Log.NewEvent("lc expired",
			glog.LogElementEvent,
			-1,
		).
			Write("aura", "lc").
			Write("target", r.self.Key()).
			Write("hydro", r.Durability[info.ReactionModKeyHydro]).
			Write("electro", r.Durability[info.ReactionModKeyElectro])
	}
}

func (r *Reactable) nextTickLC(src int) func() {
	return func() {
		if r.ecTickSrc != src {
			// source changed, do nothing
			return
		}
		// ec SHOULD be active still, since if not we would have
		// called cleanup and set source to -1
		if r.Durability[info.ReactionModKeyElectro] < info.ZeroDur || r.Durability[info.ReactionModKeyHydro] < info.ZeroDur {
			return
		}

		// update dmg based on current contributors
		r.ecAtk.FlatDmg = r.calcFinalLCDmg(r.ecAtk)

		// so ec is active, which means both aura must still have value > 0; so we can do dmg
		r.core.QueueAttackWithSnap(
			r.ecAtk,
			r.ecSnapshot,
			combat.NewSingleTargetHit(r.self.Key()),
			0,
		)

		// queue up next tick
		r.core.Tasks.Add(r.nextTickLC(src), 120)
	}
}

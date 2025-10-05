package reactable

import (
	//"fmt"

	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

// if character applies hydro or electro add them to contributors
// returns true if contributor was added or updated
func (r *Reactable) AddLCContributor(a *info.AttackEvent) bool {
	// if the atk has no source char then do nothing
	// print("+++\n")
	// fmt.Print(a.Info)

	if a.Info.ActorIndex < 0 {
		return false
	}
	// if the atk is not hydro or electro then do nothing
	if a.Info.Element != attributes.Hydro && a.Info.Element != attributes.Electro {
		return false
	}
	// if the atk has no durability then do nothing
	if a.Info.Durability < info.ZeroDur {
		return false
	}
	// if there's still frozen left don't try to lc
	// game actively rejects lc reaction if frozen is present
	if r.Durability[info.ReactionModKeyFrozen] > info.ZeroDur {
		return false
	}

	// if char already exists, update their expiry frame
	// contributor expires after 360 frames (6 seconds)

	for frame, character := range r.contributors {
		if character.ActorIndex == a.Info.ActorIndex {
			delete(r.contributors, frame)
			r.contributors[r.core.F+360] = a.Info
			return true
		}
	}

	// if we have space, add new contributor
	if len(r.contributors) < 4 {
		// print("adding new contributor\n")
		r.contributors[r.core.F+360] = a.Info
		return true
	}

	// otherwise evict the oldest contributor and add new one
	return false
}

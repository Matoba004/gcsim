package dahlia

import (
	"fmt"

	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

var (
	attackFrames          [][]int
	attackHitmarks        = [][]int{{15}, {13}, {14, 15}, {21}}
	attackHitlagHaltFrame = [][]float64{{0.03}, {0.03}, {0.06, 0.06}, {0.09}}
	attackHitboxes        = [][]float64{{1.5}, {1.7}, {1.7, 1.5}, {2.2}}
	attackOffsets         = [][]float64{{0.5}, {0.5}, {0.3, 0.4}, {0.6}}
)

const normalHitNum = 4

func init() {
	attackFrames = make([][]int, normalHitNum)

	attackFrames[0] = frames.InitNormalCancelSlice(attackHitmarks[0][0], 32) // N1 -> W
	attackFrames[0][action.ActionAttack] = 21                                // N1 -> N2
	attackFrames[0][action.ActionCharge] = 21                                // N1 -> CA

	attackFrames[1] = frames.InitNormalCancelSlice(attackHitmarks[1][0], 34) // N2 -> W
	attackFrames[1][action.ActionAttack] = 21                                // N2 -> N3
	attackFrames[1][action.ActionCharge] = 21                                // N2 -> CA

	attackFrames[2] = frames.InitNormalCancelSlice(attackHitmarks[2][1], 47) // N3 -> W
	attackFrames[2][action.ActionAttack] = 39                                // N3 -> N4
	attackFrames[2][action.ActionCharge] = 39                                // N3 -> CA

	attackFrames[3] = frames.InitNormalCancelSlice(attackHitmarks[3][0], 62) // N4 -> W
	attackFrames[3][action.ActionAttack] = 71                                // N4 -> N1
	attackFrames[3][action.ActionCharge] = 500                               // TODO: this action is illegal; need better way to handle it
}

func (c *char) Attack(p map[string]int) (action.Info, error) {
	for i, mult := range attack[c.NormalCounter] {
		ai := info.AttackInfo{
			ActorIndex:         c.Index(),
			Abil:               fmt.Sprintf("Normal %v", c.NormalCounter),
			AttackTag:          attacks.AttackTagNormal,
			ICDTag:             attacks.ICDTagNormalAttack,
			ICDGroup:           attacks.ICDGroupDefault,
			StrikeType:         attacks.StrikeTypeSlash,
			Element:            attributes.Physical,
			Durability:         25,
			Mult:               mult[c.TalentLvlAttack()],
			HitlagFactor:       0.01,
			HitlagHaltFrames:   attackHitlagHaltFrame[c.NormalCounter][i] * 60,
			CanBeDefenseHalted: true,
		}
		ap := combat.NewCircleHitOnTarget(
			c.Core.Combat.Player(),
			info.Point{Y: attackOffsets[c.NormalCounter][i]},
			attackHitboxes[c.NormalCounter][i],
		)
		c.QueueCharTask(func() {
			c.Core.QueueAttack(ai, ap, 0, 0)
		}, attackHitmarks[c.NormalCounter][i])
	}

	defer c.AdvanceNormalIndex()

	return action.Info{
		Frames:          frames.NewAttackFunc(c.Character, attackFrames),
		AnimationLength: attackFrames[c.NormalCounter][action.InvalidAction],
		CanQueueAfter:   attackHitmarks[c.NormalCounter][len(attackHitmarks[c.NormalCounter])-1],
		State:           action.NormalAttackState,
	}, nil
}

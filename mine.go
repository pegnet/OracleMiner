package OracleMiner

import (
	"github.com/FactomProject/factomd/common/primitives/random"
	"github.com/pegnet/LXR256"
)


type Mine struct {
	Control        chan int    // sending any int to the Mine will stop mining
	OPR            [] byte     // The oracle Record that we were mining
	OprHash        []byte      // The hash of the oracle record
	BestDifficulty uint64      // Highest Difference Found
	BestNonce      []byte      // The Nonce that produced the bestDifference
	BestHash       []byte      // The best hash we found (to check the Best Difficulty against)
	lx             lxr.LXRHash // The lxr hash struct
}

func GetFirstNonce() []byte {
	// Start with 8 bytes of zero
	nonce := []byte {0,0,0,0,0,0,0,0}
	// Pick a random number to fill out our nonce (don't collide with other miners if in a pool)
	nonce = append(nonce,random.RandByteSliceOfLen(24)...)
	return nonce
}

// Mine a given Oracle Price Record (OPR)
func (m *Mine) Mine(opr []byte) {
	var lx lxr.LXRHash
	lx.Init()

	m.OPR = opr
	m.OprHash = lx.Hash(opr)

	// Clear out some variables in case the Mine struct is being reused
	m.BestDifficulty = 0
	m.BestNonce = m.BestNonce[:0]
	m.BestHash = m.BestHash[:0]

	// Put my nonce and Opr together... We hash them both
	nonceOpr := GetFirstNonce()
	nonceOpr = append(nonceOpr,m.OprHash...)

	for {

		// The process ends when signaled.
		select {
		case <-m.Control:
			break
		default:
		}

		// Increment my nonce
				for i := 0; ; i++ {
					nonceOpr[i] += 1
					if nonceOpr[i] != 0 {
						break
					}
				}

				try := lx.Hash(nonceOpr)

				d := lxr.Difficulty(try)
				if d == 0 {
					continue
				}
				if m.BestDifficulty == 0 || m.BestDifficulty > d {
					m.BestDifficulty = d
					m.BestNonce = append(m.BestNonce[:0], nonceOpr[:32]...)
				}
			}


	}



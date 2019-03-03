package OracleMiner

import (
	"github.com/FactomProject/factomd/common/primitives/random"
	"github.com/PaulSnow/LXR256"
)

type Mine struct {
	Finished       bool     // Miner sets Finished with the process is killed.
	Control        chan int // sending any int to the Mine will stop mining
	OPR            []byte   // The oracle Record that we were mining
	OprHash        []byte   // The hash of the oracle record
	BestDifficulty uint64   // Highest Difference Found
	BestNonce      []byte   // The Nonce that produced the bestDifference
	BestHash       []byte   // The best hash we found (to check the Best Difficulty against)
	Hashcnt        int      // Count of hash rounds performed.
}

func GetFirstNonce() []byte {
	// Start with 8 bytes of zero
	nonce := []byte{0, 0, 0, 0, 0, 0, 0, 0}
	// Pick a random number to fill out our nonce (don't collide with other miners if in a pool)
	nonce = append(nonce, random.RandByteSliceOfLen(24)...)
	return nonce
}

type hashFunction func([]byte) []byte

// Mine a given Oracle Price Record (OPR)
func (m *Mine) Mine(hash hashFunction, opr []byte) {

	m.OPR = opr
	m.OprHash = hash(opr)
	m.Hashcnt = 0

	// Clear out some variables in case the Mine struct is being reused
	m.BestDifficulty = 0
	m.BestNonce = m.BestNonce[:0]
	m.BestHash = m.BestHash[:0]

	// Put my nonce and Opr together... We hash them both
	nonceOpr := GetFirstNonce()
	nonceOpr = append(nonceOpr, m.OprHash...)

	for {

		// The process ends when signaled.
		select {
		case <-m.Control:
			m.Finished = true
			return
		default:
		}

		// Increment my nonce
		for i := 0; ; i++ {
			nonceOpr[i] += 1
			if nonceOpr[i] != 0 {
				break
			}
		}

		m.Hashcnt++
		try := hash(nonceOpr)

		d := lxr.Difficulty(try)
		if d == 0 {
			continue
		}
		if m.BestDifficulty == 0 || m.BestDifficulty > d {
			m.BestDifficulty = d
			m.BestNonce = append(m.BestNonce[:0], nonceOpr[:32]...)
			m.BestHash = append(m.BestHash[:0], try...)
		}
	}
}

const (
	startofblock = iota
	blocktime
	minute8
	faulting
)

type FactomEvent struct {
	Type  int
	Value int
	Msg   string
}

func (m *Mine) Events(event chan FactomEvent) {

}

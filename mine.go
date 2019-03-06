package OracleMiner

import (
	"github.com/FactomProject/factomd/common/primitives/random"
	"github.com/pegnet/LXR256"
)

type hashFunction func([]byte) []byte

type Mine struct {
	Response       chan int // Returns 0 when the miner stops
	Control        chan int // sending any int to the Mine will stop mining
	OPR            []byte   // The oracle Record that we were mining
	OprHash        []byte   // The hash of the oracle record
	BestDifficulty uint64   // Highest Difference Found
	BestNonce      []byte   // The Nonce that produced the bestDifference
	BestHash       []byte   // The best hash we found (to check the Best Difficulty against)
	Hashcnt        int      // Count of hash rounds performed.
	HashFunction   hashFunction // The Hash function we will be using to mine
}

func (m *Mine) Init(){
	m.BestDifficulty = 0
	m.Hashcnt = 0
	m.Response = make(chan int, 10)
	m.Control = make(chan int, 10)
	m.OPR = m.OPR[:0]
	m.OprHash = m.OPR[:0]
	m.BestNonce = m.OPR[:0]
	m.BestHash = m.OPR[:0]

	// create an LXRHash function for the Lookup XoR hash.
	lx := new(lxr.LXRHash)
	lx.Init()
	m.HashFunction = func(src []byte) []byte {
		return lx.Hash(src)
	}
}

// Start mining on a given OPR
func (m *Mine) Start(opr [] byte) (){
	// Make sure the miner is stopped.
	if m.Response != nil {
		m.Stop()
	}
	m.Init()
	go m.Mine(opr)
}

func (m *Mine) Stop() {
	// Only stop a running miner
	if m.Response == nil {
		// Signal the mining process to stop
		m.Control <- 0
		// Wait for it to stop
		<-m.Response
		// Clear the response channel to indicate the miner is stopped.
		m.Response = nil
	}
}


// Create a nonce of eight bytes of zeros followed by 24 bytes of random values
func GetFirstNonce() []byte {
	// Start with 8 bytes of zero
	nonce := []byte{0, 0, 0, 0, 0, 0, 0, 0}
	// Pick a random number to fill out our nonce (don't collide with other miners if in a pool)
	nonce = append(nonce, random.RandByteSliceOfLen(24)...)
	return nonce
}

// Mine a given Oracle Price Record (OPR)
func (m *Mine) Mine( opr []byte) {

	m.OPR = opr
	m.OprHash = m.HashFunction(opr)
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
			m.Response <- 0
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
		try := m.HashFunction(nonceOpr)

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

package main

import (
	"fmt"
	"github.com/pegnet/OracleMiner"
	"github.com/pegnet/OracleRecord"
	"time"
	"sync"
)

const speedlimit = 600 // Don't hit the pricing APIs faster than once every 10 minutes.

type OPRState struct {
	miner int
	opr   oprecord.OraclePriceRecord
}
var	lastopr  []byte // Last OPR record
var	lasttime int64  // Time of last API call
var mutex sync.Mutex

// GetOPR
// To preserve our free access to the APIs we are using, don't actually build the OPR record quicker
// than the speedlimit.  Faster calls just get the last OPR record
func GetOPR(state *OPRState) []byte {
	mutex.Lock()
	defer mutex.Unlock()
	now := time.Now().Unix()
	if now-lasttime < 600 {
		return lastopr
	}
	lasttime = now
	state.opr.GetOPRecord()
	data, err := state.opr.MarshalBinary()
	if err != nil {
		panic("Could not produce an oracle record")
	}
	lastopr = data
	fmt.Println(state.opr.String())
	return data
}

func RunMiner(minerNumber int) {
	state := new(OPRState)
	state.miner = minerNumber



	miner := new(OracleMiner.Mine)
	miner.Init()

	fm := new(OracleMiner.FactomdMonitor)
	var blocktime int64
	alert := fm.Start()

	OracleMiner.InitNetwork(minerNumber, &state.opr)

	started := false
	for {
		min := <-alert
		block := <-alert
		fmt.Printf("Blockheight %d Minute %2d\n", block, min)
		switch min {
		case 1:
			if started == false {
				OracleMiner.GradeLastBlock(&state.opr, int64(block), miner)
				blocktime = fm.GetBlockTime()
				opr := GetOPR(state)
				miner.Start(opr)
				fmt.Println("mining started")
				started = true
			}
		case 8:
			if started {
				// sleep for half a block time.
				time.Sleep(time.Duration(int(blocktime)/10) * time.Second)
				miner.Stop()
				fmt.Printf("miner %5d Difficulty %8x Nonce %x Hash %x \n", state.miner, miner.BestDifficulty, miner.BestNonce, miner.BestHash)
				started = false
				OracleMiner.AddOpr(&state.opr, miner.BestNonce)
			}
		}

	}
}

func main() {
	for i:=0; i<100; i++ {
		go RunMiner(i + 1)
		time.Sleep(1 * time.Second)
	}
	for { time.Sleep(1*time.Second)}
}

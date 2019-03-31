package main

import (
	"encoding/binary"
	"fmt"
	"github.com/pegnet/OracleMiner"
	"time"
)

// GetOPR
// To preserve our free access to the APIs we are using, don't actually build the OPR record quicker
// than the speedlimit.  Faster calls just get the last OPR record
func GetOPR(state *OracleMiner.MinerState) []byte {
    binary.BigEndian.PutUint64(state.OPR.BlockReward[:],5000 *100000000)
	state.OPR.GetOPRecord(state.Config)
	data, err := state.OPR.MarshalBinary()
	if err != nil {
		panic("Could not produce an oracle record")
	}
	fmt.Println(state.OPR.String())
	return data
}

func RunMiner(minerNumber int) {
	fmt.Print(" ",minerNumber)
	mstate := new(OracleMiner.MinerState)
	mstate.MinerNumber = minerNumber

	mstate.MinerNumber = minerNumber
	miner := new(OracleMiner.Mine)
	miner.Init()

	fm := new(OracleMiner.FactomdMonitor)
	var blocktime int64
	alert := fm.Start()

	mstate.LoadConfig()
	OracleMiner.InitNetwork(mstate, minerNumber, &mstate.OPR)

	started := false
	for {
		min := <-alert
		block := <-alert
		switch min {
		case 1:
			if started == false {
				OracleMiner.GradeLastBlock(mstate, &mstate.OPR, int64(block), miner)
				blocktime = fm.GetBlockTime()
				opr := GetOPR(mstate)
				miner.Start(opr)
				started = true
			}
		case 8:
			if started {
				// sleep for half a block time.
				time.Sleep(time.Duration(int(blocktime)/10) * time.Second)
				miner.Stop()
				started = false
				OracleMiner.AddOpr(mstate,  miner.BestNonce)
			}
		}

	}
}

func main() {
	for i := 0; i < 60; i++ {
		go RunMiner(i + 1)
		time.Sleep(500 * time.Millisecond)
	}
	for {
		time.Sleep(1 * time.Second)
	}
}

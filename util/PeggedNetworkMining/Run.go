package main

import (
	"encoding/binary"
	"github.com/pegnet/OracleMiner"
	"math/rand"
	"time"
	"fmt"
)

// GetOPR
// To preserve our free access to the APIs we are using, don't actually build the OPR record quicker
// than the speedlimit.  Faster calls just get the last OPR record
func GetOPR(state *OracleMiner.MinerState) []byte {
	binary.BigEndian.PutUint64(state.OPR.BlockReward[:], 5000*100000000)
	state.OPR.GetOPRecord(state.Config)
	data, err := state.OPR.MarshalBinary()
	if err != nil {
		panic("Could not produce an oracle record")
	}
	return data
}

func RunMiner(minerNumber int) {
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

	funding := false
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
				funding = true
			}
		case 8:
			if started {
				// sleep for half a block time.
				time.Sleep(time.Duration(int(blocktime)/10) * time.Second)
				miner.Stop()
				started = false
				copy (mstate.OPR.Nonce[:],miner.BestNonce)
				if mstate.OPR.ComputeDifficulty() > 0 {
					OracleMiner.AddOpr(mstate, miner.BestNonce)
				}else{
					fmt.Println("miner ", mstate.MinerNumber,":  \"Man, didn't find a solution! Drat!\"")
				}
			}
		case 0:
			if funding {
				if rand.Intn(100) > 95 {
					OracleMiner.FundWallet(mstate)
				}
				funding = false
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

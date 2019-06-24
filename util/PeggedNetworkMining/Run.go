package main

import (
	"fmt"
	"github.com/pegnet/OracleMiner"
	"math/rand"
	"time"
	"github.com/FactomProject/btcutil/base58"
	"encoding/json"
)

// GetOPR
// To preserve our free access to the APIs we are using, don't actually build the OPR record quicker
// than the speedlimit.  Faster calls just get the last OPR record
func GetOPR(dbht int32, state *OracleMiner.MinerState) []byte {
	state.OPR.GetOPRecord(state.Config)
	state.OPR.Dbht = dbht
	data, err := json.Marshal(state.OPR)
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
	miner.MinerNum = minerNumber
	miner.Init()

	mstate.Monitor = new(OracleMiner.FactomdMonitor)

	var blocktime int64
	alert := mstate.Monitor.Start()
	_ = blocktime
	mstate.LoadConfig()
	OracleMiner.InitNetwork(mstate, minerNumber, &mstate.OPR)

	funding := false
	started := false
	for {
		min := <-alert
		block := <-alert
		switch min {
		case 1:
			miner.Dbht = int32(block + 1)
			if started == false {
				OracleMiner.GradeLastBlock(mstate, &mstate.OPR, int64(block), miner)
				blocktime = mstate.Monitor.GetBlockTime()
				opr := GetOPR(int32(block+1), mstate)
				miner.Start(opr)
				started = true
				funding = true
			}
		case 9:
			if started {
				// sleep for half a block time.
				miner.Stop()
				started = false
				mstate.OPR.Nonce = base58.Encode(miner.BestNonce)
				if mstate.OPR.ComputeDifficulty() > 0 {
					OracleMiner.AddOpr(mstate)
					fmt.Println("miner ", mstate.MinerNumber, ": Solution")
				} else {
					fmt.Println("miner ", mstate.MinerNumber, ": No Solution")
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
	for i := 0; i < 40; i++ {
		go RunMiner(i + 1)
		time.Sleep(100 * time.Millisecond)
	}
	for {
		time.Sleep(1 * time.Second)
	}
}

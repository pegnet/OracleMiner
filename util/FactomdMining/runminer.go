package main

import (
	"fmt"
	"github.com/pegnet/OracleMiner"
	"github.com/FactomProject/factomd/common/primitives/random"
	"time"
)


// Stub for the Oracle Price Record (OPR)
func GetOPR() []byte {
	return random.RandByteSliceOfLen(300)
}


func main() {

	miner := new(OracleMiner.Mine)

	fm := new(OracleMiner.FactomdMonitor)
	alert := make(chan int,100)
	var blocktime int64
	fm.Start(&alert)

	fm.FundWallet()

	started := false
	for {
		min := <- alert
		block := <- alert
		fmt.Printf("Blockheight %d Minute %2d\n",block, min)
		switch min {
		case 0:
			if started == false {
				opr := GetOPR()
				miner.Start(opr)
				fmt.Println ("mining started")
				started = true
			}
		case 1:
			blocktime = fm.GetBlockTime()
		case 9:
			if started {
				// sleep for half a block time.
				time.Sleep(time.Duration(int(blocktime)/10)*time.Second)
				miner.Stop()
				fmt.Println("mining stopped")
				fmt.Printf("Difficulty %8x Nonce %x Hash %x \n", miner.BestDifficulty, miner.BestNonce, miner.BestHash)
				started = false
			}
		}


	}

}


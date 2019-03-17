package main

import (
	"fmt"
	"github.com/pegnet/OracleMiner"
	"github.com/pegnet/OracleRecord"
	"time"
)

const speedlimit = 600 // Don't hit the pricing APIs faster than once every 10 minutes.

var opr oprecord.OraclePriceRecord
var lastopr []byte // Last OPR record
var lasttime int64 // Time of last API call

// GetOPR
// To preserve our free access to the APIs we are using, don't actually build the OPR record quicker
// than the speedlimit.  Faster calls just get the last OPR record
func GetOPR() []byte {
	now := time.Now().Unix()
	if now-lasttime < 600 {
		return lastopr
	}
	lasttime = now
	opr.GetOPRecord()
	data, err := opr.MarshalBinary()
	if err != nil {
		panic("Could not produce an oracle record")
	}
	lastopr = data
	fmt.Println(opr.String())
	return data
}

func main() {

	miner := new(OracleMiner.Mine)

	fm := new(OracleMiner.FactomdMonitor)
	var blocktime int64
	alert := fm.Start()

	OracleMiner.InitNetwork(fm)

	started := false
	for {
		min := <-alert
		block := <-alert
		fmt.Printf("Blockheight %d Minute %2d\n", block, min)
		switch min {
		case 0:
			if started == false {
				opr := GetOPR()
				miner.Start(opr)
				fmt.Println("mining started")
				started = true
			}
		case 1:
			blocktime = fm.GetBlockTime()
		case 9:
			if started {
				// sleep for half a block time.
				time.Sleep(time.Duration(int(blocktime)/10) * time.Second)
				miner.Stop()
				fmt.Printf("Difficulty %8x Nonce %x Hash %x \n", miner.BestDifficulty, miner.BestNonce, miner.BestHash)
				started = false
			}
		}

	}

}

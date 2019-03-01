package main

import (
	"fmt"
	"github.com/FactomProject/factomd/common/primitives/random"
	"github.com/pegnet/OracleMiner"
	"time"
)

func main() {

	kill := make(chan int)
	miner := new(OracleMiner.Mine)
	miner.Control = kill

	for {
		// Get something like a 300 byte OPR record
		opr := random.RandByteSliceOfLen(300)
		go miner.Mine(opr)
		time.Sleep(10 * time.Second)
		miner.Control <- 1
		fmt.Printf("%30s %x\n", "mined the OPR", opr[:32])
		fmt.Printf("%30s %x\n", "difficulty", miner.OprHash)
		fmt.Printf("%30s %x\n", "nonce", miner.BestNonce)
		fmt.Printf("%30s %x\n", "difficulty", miner.BestDifficulty)
		fmt.Println()
	}

}

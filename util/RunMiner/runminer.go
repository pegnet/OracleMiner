package main

import (
	"fmt"
	"github.com/FactomProject/factomd/common/primitives/random"
	"github.com/dustin/go-humanize"
	"github.com/pegnet/LXR256"
	"github.com/pegnet/OracleMiner"
	"time"
)

func main() {

	// Constants to control this example mining process.
	const NumberMiners = 5        // Number of mining processes to launch
	blocktime := 10 * time.Second // blocktime to mine

	var miners [NumberMiners]*OracleMiner.Mine

	fmt.Printf("Running %d miners with a blocktime of %d seconds\n\n", NumberMiners, blocktime/time.Second)

	for i := range miners {
		// Setup the mining process
		kill := make(chan int)
		miners[i] = new(OracleMiner.Mine)
		miners[i].Control = kill
	}

	// create an LXRHash function for the Lookup XoR hash.
	lx := new(lxr.LXRHash)
	lx.Init()
	hashFunction := func(src []byte) []byte {
		return lx.Hash(src)
	}
	blk := 0
	// As a test function, we simply create a 300 byte random value buffer as a standin for the OPR record.
	for {
		// Get something like a 300 byte OPR record
		opr := random.RandByteSliceOfLen(300)
		blk++

		// Start the mining process on the record using a range of miners
		for _, miner := range miners {
			go miner.Mine(hashFunction, opr)
		}

		time.Sleep(blocktime)

		fmt.Println("=========================================================================")
		for i, miner := range miners {
			// stop each miner
			miner.Control <- 1

			for !miner.Finished {
				time.Sleep(10 * time.Millisecond)
			}
			fmt.Printf("Miner %3d block %6d hash/sec %s\n",
				i+1, blk, humanize.Comma(int64(miner.Hashcnt)/int64(blocktime.Seconds())))
			fmt.Printf("%30s %x\n", "mined the OPR", opr[:32])
			fmt.Printf("%30s %x\n", "difficulty", miner.OprHash)
			fmt.Printf("%30s %x\n", "nonce", miner.BestNonce)
			fmt.Printf("%30s %x\n", "best hash", miner.BestHash)
			fmt.Printf("%30s %x\n", "difficulty", miner.BestDifficulty)
			fmt.Println()
		}
	}

}

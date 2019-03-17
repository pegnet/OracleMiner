package OracleMiner

import (
	"fmt"
	"github.com/pegnet/LXR256"
)

type OPR struct {
	ID                 int        // not part of OPR -an index for debugging
	Grade              float64    // not part of OPR -grade for this OPR relative to the average
	Difficulty         uint64     // not part of OPR -The difficulty of the given nonce
	Designators        [20]string // not part of OPR -(USD, YEN, etc.) pulled from PegomChain
	OprHash            []byte     // Hash of the OPRbinary
	Nonce              [32]byte   // Provided by the miner on chain
	OPRbinary          []byte     // The binary OPR record
	PegomChainID       [32]byte   // Chain ID of the tokens defined by Pegom
	WinningOPR         [32]byte   // Winning OPR of the last round
	CoinbasePNTaddress [32]byte   // Coinbase address to recieve the reward if paid
	Reward             int64      // Total reward issued in the round
	FactomDID          [32]byte   // Factom Identity for the miner
	Tokens             [20]int64  // A list of all the balances of all the tracked tokens in order
}

func (o *OPR) String() string {
	return fmt.Sprintf("Grade %40.0f Difficulty %16x Nonce %30x", o.Grade, o.Difficulty, o.Nonce[:14])
}

// Compute the average answer for the price of each token reported
func Avg(list []*OPR) (avg [20]float64) {
	// Sum up all the prices
	for _, opr := range list {
		for i, price := range opr.Tokens {
			avg[i] += float64(price)
		}
	}
	// Then divide the prices by the number of OPR records.  Two steps is actually faster
	// than doing everything in one loop (one divide for every asset rather than one divide
	// for every asset * number of OPRs)
	numList := float64(len(list))
	for i := range avg {
		avg[i] = avg[i] / numList / 100000000
	}
	return
}

// Given the average answers across a set of tokens, grade the opr
func CalculateGrade(avg [20]float64, opr *OPR) float64 {
	for i, v := range opr.Tokens {
		d := float64(v)/100000000 - avg[i] // compute the difference from the average
		opr.Grade = opr.Grade + d*d*d*d    // the grade is the sum of the squares of the differences
	}
	return opr.Grade
}

// Given a list of OPR, figure out which 10 should be paid, and in what order
func grade(list []*OPR) (tobepaid []*OPR, sortedlist []*OPR) {
	lx := lxr.LXRHash{}
	lx.Init()
	if len(list) <= 10 {
		return nil, nil
	}

	// Calculate the difficult for each entry in the list of OPRs.
	for _, opr := range list {
		// append the Nounce and OPR ==> no
		no := []byte{}
		no = append(no, opr.Nonce[:]...)  // Get the nonce (32 bytes)
		oprHash := lx.Hash(opr.OPRbinary) // get the hash of the opr (32 bytes)
		no = append(no, oprHash...)       // append the opr hash
		h := lx.Hash(no)                  // we hash the 64 resulting bytes.

		opr.Difficulty = lxr.Difficulty(h) // Go calculate the difficulty, and cache in the opr
	}
	last := len(list)
	// Throw away all the entries but the top 50 in difficulty
	if len(list) > 50 {
		// bubble sort because I am lazy.  Could be replaced with about anything
		for j := 0; j < len(list)-1; j++ {
			for k := 0; k < len(list)-j-1; k++ {
				if list[k].Difficulty > list[k+1].Difficulty { // sort the largest difficulty to the end of the list
					list[k], list[k+1] = list[k+1], list[k]
				}
			}
		}
		last = 50 // truncate the list to the best 50
	}

	// Go through and throw away entries that are outside the average or on a tie, have the worst difficulty
	// until we are only left with 10 entries to reward
	for i := last; i >= 10; i-- {
		avg := Avg(list[:i])
		for j := 0; j < i; j++ {
			CalculateGrade(avg, list[j])
		}
		// bubble sort the worst grade to the end of the list. Note that this is nearly sorted data, so
		// a bubble sort with a short circuit is pretty darn good sort.
		for j := 0; j < i-1; j++ {
			cont := false                // If we can get through a pass with no swaps, we are done.
			for k := 0; k < i-j-1; k++ { // yes, yes I know we can get 2 or 3 x better speed playing with indexes
				if list[k].Grade > list[k+1].Grade { // bit it is tricky.  This is good enough.
					list[k], list[k+1] = list[k+1], list[k] // sort first by the grade.
					cont = true                             // any swap means we continue to loop
				} else if list[k].Grade == list[k+1].Grade { // break ties with PoW.  Where data is being shared
					if list[k].Difficulty > list[k+1].Difficulty { // we will have ties.
						//list[k], list[k+1] = list[k+1], list[k]
						cont = true // any swap means we continue to loop
					}
				}
			}
			if !cont { // If we made a pass without any swaps, we are done.
				break
			}
		}
	}
	tobepaid = append(tobepaid, list[:10]...)
	return tobepaid, list
}

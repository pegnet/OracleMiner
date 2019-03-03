package OracleMiner

import (
	"fmt"
	"github.com/FactomProject/factomd/common/primitives/random"
	"github.com/pegnet/LXR256"
	"math/rand"
	"testing"
	"time"
)

func TestGrading(t *testing.T) {

	lx := new(lxr.LXRHash)
	lx.Init()
	hashFunction := func(src []byte) []byte {
		return lx.Hash(src)
	}

	var miners [16]*Mine
	for i := range miners {
		kill := make(chan int)
		miners[i] = new(Mine)
		miners[i].Control = kill
	}

	totaltime := int64(0)

	// Run lots of tests.
	for i := 0; i < 1000000; i++ {
		// create an average price list
		avg := [20]float64{}
		for i := range avg {
			avg[i] = rand.Float64() * 1000
		}

		// create 30 to 200 test OPRs
		n := rand.Intn(20) + 50
		fmt.Println("We will be testing ", n, "OPR records")
		oprList := []*OPR{}
		for i := 0; i < n; i++ {
			opr := new(OPR)
			oprList = append(oprList, opr)
			opr.OPRbinary = random.RandByteSliceOfLen(300)
			for i, v := range avg {
				t := int64((v + v*rand.Float64()/99) * 100000000)
				opr.Tokens[i] = t
			}
		}

		goodenough := 0
		// make a list of all the OPR that are not yet mined.
		notmined := []*OPR{}
		notmined = append(notmined, oprList...)

		// Mine some difficulty for all our sample data
		for goodenough < n {
			fmt.Println("goodenough ", goodenough, " n ", n)
			started := 0
			for i, m := range miners {

				if i >= len(notmined) {
					break
				}
				started++
				notmined[i].OprHash = lx.Hash(notmined[i].OPRbinary)
				go m.Mine(hashFunction, notmined[i].OPRbinary)
			}
			time.Sleep(1 * time.Second)
			for i, m := range miners {
				if i >= started {
					break
				}
				m.Control <- 1
				for !m.Finished {
					time.Sleep(10 * time.Millisecond)
				}
				if m.BestDifficulty > 0 {
					copy(notmined[i].Nonce[:], m.BestNonce)
					notmined[i] = nil
					goodenough++
				}
			}
			notmined2 := []*OPR{}
			for _, m := range notmined {
				if m != nil {
					notmined2 = append(notmined2, m)
				}
			}
			notmined = notmined2
		}

		// now grade the results

		start := time.Now().UnixNano()
		reward, order := grade(oprList)
		end := time.Now().UnixNano()
		totaltime = totaltime + end - start

		// Some debugging for the grading.
		if true {
			fmt.Println("Top 10")
			for i, r := range reward {
				fmt.Printf("%4d %s\n", i, r.String())
			}
			fmt.Println("Graded List")
			for i, o := range order {
				fmt.Printf("%4d %s\n", i, o.String())
			}
			fmt.Printf("Time to run on average: %v microseconds\n", totaltime/1000/int64(i+1))
		}
	}
}

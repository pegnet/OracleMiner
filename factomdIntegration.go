package OracleMiner

import (
	"fmt"
	"github.com/FactomProject/factom"
	"sync"
	"time"
)

// FactomdMonitor
// Running multiple Monitors is problematic and should be avoided if possible
type FactomdMonitor struct {
	mutex                   sync.Mutex // Protect multiple parties accessing monitor data
	lastminute              int64      // Last minute we got
	lastblock               int64      // Last block we got
	polltime                int64      // How frequently do we poll
	kill                    chan int   // Channel to kill polling.
	response                chan int   // Respond when we have stopped
	alert                   chan int   // Channel to send minutes to
	polls                   int64
	leaderheight            int64
	directoryblockheight    int64
	minute                  int64
	currentblockstarttime   int64
	currentminutestarttime  int64
	currenttime             int64
	directoryblockinseconds int64
	stalldetected           bool
	faulttimeout            int64
	roundtimeout            int64
	status                  string
	fcts                    []*factom.FactoidAddress
	ecs                     []*factom.ECAddress
}

// GetBlockTime
// Returns the blocktime in seconds.  All blocks are divided into 10 "minute" sections.  But if the blocktime
// is not 600 seconds, then a minute = the blocktime/10
func (f *FactomdMonitor) GetBlockTime() int64 {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	return f.directoryblockinseconds
}

// poll
// Go process to poll the Factoid client to provide insight into its operations.
func (f *FactomdMonitor) poll() {
	fmt.Println("Start Factomd Polling")
	defer func() { fmt.Println("Stop Factomd Polling") }()
	for {
		var err error
		for {
			f.mutex.Lock()

			// If we have a kill message, die!
			select {
			case <-f.kill:
				f.response <- 1
				f.mutex.Unlock()
				return
			default:
			}

			// Do our poll
			f.leaderheight,
				f.directoryblockheight,
				f.minute,
				f.currentblockstarttime,
				f.currentminutestarttime,
				f.currenttime,
				f.directoryblockinseconds,
				f.stalldetected,
				f.faulttimeout,
				f.roundtimeout,
				err = factom.GetCurrentMinute()

			// track how often we poll
			f.polls++

			// If we get an error, then report and break
			if err != nil {
				f.status = err.Error()
				panic("Error with getting minute.")

				break
			}
			// If we got a different block time, consider that good and break
			if f.minute != f.lastminute || f.directoryblockheight != f.lastblock {
				f.lastminute = f.minute
				f.lastblock = f.directoryblockheight
				break
			}

			// Poll once per second until we get a new minute
			f.mutex.Unlock()
			time.Sleep(1 * time.Second)
		}

		f.alert <- int(f.minute)
		f.alert <- int(f.directoryblockheight)
		f.mutex.Unlock()
		// Poll once per second
		time.Sleep(time.Duration(time.Second))
	}
}

func (f *FactomdMonitor) Start() chan int {
	f.mutex.Lock()
	if f.kill == nil {
		f.response = make(chan int, 1)
		f.alert = make(chan int, 10)
		f.kill = make(chan int, 1)
		factom.SetFactomdServer("localhost:8088")
		factom.SetWalletServer("localhost:8089")
		go f.poll()
	}
	f.mutex.Unlock()
	return f.alert
}

func (f *FactomdMonitor) Stop() {
	f.mutex.Lock()
	f.kill <- 0
	<-f.response
	f.mutex.Unlock()
}

// FundWallet()
// This is just a debugging function.  These addresses work when run against a LOCAL network simulation.
func (f *FactomdMonitor) FundWallet() (err error) {
	f.fcts, f.ecs, err = factom.ImportAddresses("Fs3E9gV6DXsYzf7Fqx1fVBQPQXV695eP3k5XbmHEZVRLkMdD9qCK", "Es3LB2YW9bpdWmMnNQYb31kyPzqnecsNqmg5W4K7FKp4UP6omRTa")
	if err != nil {
		return
	}
	factom.DeleteTransaction("fundec")
	rate, err := factom.GetRate()
	if err != nil {
		return
	}
	_, err = factom.NewTransaction("fundec")
	if err != nil {
		return
	}

	fct2ec := uint64(100) * 100000000 // Buy so many FCT (shift left 8 decimal digits to create a fixed point number)
	_, err = factom.AddTransactionInput("fundec", f.fcts[0].String(), fct2ec)
	if err != nil {
		return
	}
	factom.AddTransactionECOutput("fundec", f.ecs[0].PubString(), fct2ec)
	_, err = factom.AddTransactionFee("fundec", f.fcts[0].String())
	if err != nil {
		return
	}
	_, err = factom.SignTransaction("fundec", false)
	if err != nil {
		return
	}
	_, err = factom.SendTransaction("fundec")
	if err != nil {
		return
	}

	fmt.Println("Bought: ", fct2ec/rate, " Entry Credits")

	return
}

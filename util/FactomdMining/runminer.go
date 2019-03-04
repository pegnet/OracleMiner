package main

import (
	"github.com/FactomProject/factom"
	"sync"
	"time"
	"fmt"
)

// FactomdMonitor
// Running multiple Monitors is problematic and should be avoided if possible
type FactomdMonitor struct {
	mutex      sync.Mutex // Protect multiple parties accessing monitor data
	lastminute int64      // Last minute we got
	lastblock int64      // Last block we got
	polltime   int64      // How frequently do we poll
	kill       *chan int  // Channel to kill polling.
	alert      *chan int  // Channel to send minutes to

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

	status string
}

// poll
// Go process to poll the Factoid client to provide insight into its operations.
func (f *FactomdMonitor) poll() {

	for {
		var err error
		for {
			f.mutex.Lock()

			// If we have a kill message, die!
			select {
			case <-*f.kill:
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

		*f.alert <- int(f.minute)
		*f.alert <- int(f.directoryblockheight)
		fmt.Printf("go %s Blockheight %d Minute %2d\n",f.status, f.directoryblockheight+1,f.minute)
		f.mutex.Unlock()
		// Poll once per second
		time.Sleep(time.Duration(time.Second))
	}
}

func (f *FactomdMonitor) Start(alert *chan int) {
	f.mutex.Lock()
	if f.kill == nil {
		factom.SetFactomdServer("localhost:8088")
		f.kill = new(chan int)
		f.alert = alert
		go f.poll()
	}
	f.mutex.Unlock()
}

func (f *FactomdMonitor) Stop() {
	f.mutex.Lock()
	*f.kill <- 0
	f.mutex.Unlock()
}


func main() {

	fm := new(FactomdMonitor)
	alert := make(chan int,100)

	fm.Start(&alert)
	for {
		min := <- alert
		block := <- alert
		fmt.Printf("Blockheight %d Minute %2d\n",block, min)
	}


	for { time.Sleep(time.Second)}
}

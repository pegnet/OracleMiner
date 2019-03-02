package main

import (
	"fmt"
	"github.com/FactomProject/factom"

)




func main() {

	factom.SetFactomdServer("localhost:8088")

	Leaderheight,
		Directoryblockheight,
		Minute,
		Currentblockstarttime,
		Currentminutestarttime,
		Currenttime,
		Directoryblockinseconds,
		Stalldetected,
		Faulttimeout,
		Roundtimeout,
		err :=
		factom.GetCurrentMinute()

	if err != nil {
		fmt.Println(err.Error())

	}
	// Just to make the errors go away without the documentation of what I'm getting back.
	_, _, _, _, _, _, _, _, _, _ = Leaderheight, Directoryblockheight, Minute, Currentblockstarttime,
		Currentminutestarttime, Currenttime, Directoryblockinseconds, Stalldetected,
		Faulttimeout, Roundtimeout

	fmt.Printf("Block %d Minute %d Currentblockstarttime %d\n", Directoryblockheight, Minute, Currentblockstarttime)



	a,b,_,_,_,_,_,_ := factom.GetProperties()
	fmt.Println(a,b)
	print("ha")

		r, err := factom.GetRate()
		if err != nil {
			fmt.Println("ERROR:", err)
		}
		fmt.Println(r)

}

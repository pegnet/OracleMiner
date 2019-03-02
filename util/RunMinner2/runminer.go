package main

import (
	"fmt"
	"github.com/FactomProject/factom"
	"net/http/httptest"
	"net/http"
)

func main() {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
	}))
	defer ts.Close()

	url := ts.URL[7:]
	factom.SetFactomdServer(url)

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
}

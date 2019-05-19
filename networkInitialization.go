package OracleMiner

import (
	"encoding/hex"
	"fmt"
	"github.com/FactomProject/FactomCode/common"
	"github.com/FactomProject/factom"
	"github.com/pegnet/OracleRecord"
	"math/rand"
	"time"
)

func InitNetwork(mstate *MinerState, minerNumber int, opr *oprecord.OraclePriceRecord) {

	PegNetChain := mstate.GetProtocolChain()
	BPegNetChainID, err := hex.DecodeString(PegNetChain)
	if err != nil {
		panic("Could not decode the protocol chain:" + err.Error())
	}
	opr.SetChainID(BPegNetChainID)

	bOprChainID, err := hex.DecodeString(mstate.GetProtocolChain())
	if err != nil {
		panic("No OPR Chain found in config file")
	}

	did := mstate.GetIdentityChainID()
	BFactomDigitalID, _ := hex.DecodeString(did)

	opr.SetFactomDigitalID(BFactomDigitalID)

	opr.SetVersionEntryHash(common.Sha([]byte("an entry")).Bytes())

	sECAdr := mstate.GetECAddress()
	ecAdr, err := factom.FetchECAddress(sECAdr)
	if err != nil {
		fmt.Println("Failed to initialize ", minerNumber)
		fmt.Println(err.Error())
		return
	}
	opr.EC = ecAdr

	FundWallet(mstate)

	// First check if the network has been initialized.  If it hasn't, then create all
	// the initial structures.  This is only needed while testing.
	chainid := mstate.GetProtocolChain()
	if !factom.ChainExists(chainid) {
		CreatePegNetChain(mstate)
	}
	chainid = mstate.GetOraclePriceRecordChain()
	if !factom.ChainExists(chainid) {
		CreateOPRChain(mstate)
	}

	var entries []*factom.Entry
	// Check that we have an asset entry
	for {
		entries, err = factom.GetAllChainEntries(chainid)
		if err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	// I'm doing a cheap lambda function here because we are looking for a finish, and breaking
	// if we find it.
	func() {
		for _, entry := range entries {
			// We are looking for the first Asset Entry.  If we upgrade the network
			// we will have to search for the 2nd or 3rd, or whatever.  Upgrades will
			// be determined by the wallet code and the miners.
			if len(entry.ExtIDs) > 0 && string(entry.ExtIDs[0]) == "Asset Entry" {
				// If we have the Asset Entry, set that hash
				opr.SetVersionEntryHash(entry.Hash())
				return
			}
		}
		AddAssetEntry(mstate)
	}()

	// Check for and create the Oracle Price Records Chain
	chainid = hex.EncodeToString(bOprChainID)
	if !factom.ChainExists(chainid) {
		CreateOPRChain(mstate)
	}
	return
}

func check(err error) {
	if err != nil {
		fmt.Println("If you are initializing a test network, you may need to wait a block and try again.")
		panic(err)
	}
}

// FundWallet()
// This is just a debugging function.  These addresses work when run against a LOCAL network simulation.
func FundWallet(m *MinerState) (err error) {
	// Get our EC address
	ecadr := m.GetECAddress()
	fctadr := m.GetFCTAddress()
	// Check and see if we have at least 1000 entry credits
	bal, err := factom.GetECBalance(ecadr)
	check(err)
	// If we have the 1000 entry credits, then we are happy clams, and can move on!
	if bal > 500 {
		return
	}

	factom.DeleteTransaction("fundec")
	rate, err := factom.GetRate()
	check(err)

	for i:=0;i<15 ;i++ {
		_, err = factom.NewTransaction("fundec")
		if err == nil { break }
		time.Sleep(100*time.Millisecond)
	}
	check(err)

	fct2ec := uint64(2000) * rate // Buy so many FCT (shift left 8 decimal digits to create a fixed point number)
	_, err = factom.AddTransactionInput("fundec", fctadr, fct2ec)
	check(err)

	factom.AddTransactionECOutput("fundec", ecadr, fct2ec)

	for i:=0;i<15 ;i++ {
		_, err = factom.AddTransactionFee("fundec", fctadr)
		if err == nil { break }
		time.Sleep(100*time.Millisecond)
	}
	check(err)

	_, err = factom.SignTransaction("fundec", false)
	check(err)

	_, err = factom.SendTransaction("fundec")
	check(err)

	fmt.Println("Bought: ", fct2ec/rate, " Entry Credits")

	return
}

// The Pegged Network has a defining chain.  This function builds and returns the expected defining chain
// for the network.
func CreatePegNetChain(mstate *MinerState) {
	sECAdr := mstate.GetECAddress()
	ec_adr, err := factom.FetchECAddress(sECAdr)
	check(err)

	// Create the first entry for the PegNetChain
	sPegNetChainID := mstate.GetProtocolChain()
	sPegNetChainExtIDs := mstate.GetProtocolChainExtIDs()
	e := NewEntryStr(sPegNetChainID, sPegNetChainExtIDs, "")

	pegNetChain := factom.NewChain(e)

	var txid string
	for i:=0;i<1000; i++ {
		txid, err = factom.CommitChain(pegNetChain, ec_adr)
		if err == nil {
			break
		}
		if i == 0 {
			fmt.Println("Initialization... Waiting to write PegNet chain...")
		}else{
			fmt.Print(i*5 ," seconds ")
		}
		time.Sleep(5*time.Second)
	}

	check(err)
	fmt.Println("\nCreated PegNet chain ", txid)
	factom.RevealChain(pegNetChain)
}

// The Pegged Network has a defining chain.  This function builds and returns the expected defining chain
// for the network.
func AddAssetEntry(mstate *MinerState) {
	fmt.Print("miner ",mstate.MinerNumber," ")
	// Create an entry credit address
	sECAdr := mstate.GetECAddress()
	ec_adr, err := factom.FetchECAddress(sECAdr)

	assets := []string{
		"Asset Entry",
		"PegNet,PNT,PNT",
		"US Dollar,USD,pUSD",
		"Euro,EUR,pEUR",
		"Japanese Yen,JPY,pJPY",
		"Pound Sterling,GBP,pGBP",
		"Canadian Dollar,CAD,pCAD",
		"Swiss Franc,CHF,pCHF",
		"Indian Rupee,INR,pINR",
		"Singapore Dollar,SGD,pSGD",
		"Chinese Yuan,CNY,pCNY",
		"HongKong Dollar,HKD,pHKD",
		"Gold Troy Ounce,XAU,pXAU",
		"Silver Troy Ounce,XAG,pXAG",
		"Palladium Troy Ounce,XPD,pXPD",
		"Platinum Troy Ounce,XPT,pXPT",
		"Bitcoin,XBT,pXBT",
		"Ethereum,ETH,pETH",
		"Litecoin,LTC,pLTC",
		"BitcoinCash,XBC,pXBC",
		"Factom,FCT,pFCT",
	}

	// Create the first entry for the PegNetChain
	PegNetChainID := mstate.GetProtocolChain()

	assetEntry := NewEntryStr(PegNetChainID, assets, "")
	txid := ""
	for {
		txid, err = factom.CommitEntry(assetEntry, ec_adr)
		if err == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	_ = txid
	factom.RevealEntry(assetEntry)
}

// The Pegged Network has a defining chain.  This function builds and returns the expected defining chain
// for the network.
func CreateOPRChain(mstate *MinerState) {
	// Create an entry credit address
	ec_adr, err := factom.FetchECAddress(mstate.GetECAddress())
	// Create the first entry for the OPR Chain
	oprChainID := mstate.GetOraclePriceRecordChain()
	oprExtIDs := mstate.GetOraclePriceRecordExtIDs()

	e := NewEntryStr(oprChainID, oprExtIDs, "")

	oprChain := factom.NewChain(e)

	var txid string
	for i:=0;i<1000; i++ {
		txid, err = factom.CommitChain(oprChain, ec_adr)
		if err == nil {
			break
		}
		if i == 0 {
			fmt.Println("Initialization... Waiting to write Oracle Record Chain...")
		}else{
			fmt.Print(i*5 ," seconds ")
		}
		time.Sleep(1*time.Second)
	}

	check(err)

	fmt.Println("\nCreated Oracle Price Record chain ", txid)
	factom.RevealChain(oprChain)
}

// The Pegged Network has a defining chain.  This function builds and returns the expected defining chain
// for the network.
func AddOpr(mstate *MinerState, nonce []byte) {
	opr := mstate.OPR



	// Create the OPR Entry
	// Create the first entry for the OPR Chain
	oprChainID := mstate.GetOraclePriceRecordChain()
	bOpr, err := opr.MarshalBinary()

//	assetEntry := opr.GetEntry(mstate.GetOraclePriceRecordChain())
//	_, err := factom.CommitEntry(assetEntry, opr.EC)
	check(err)




	entryExtIDs := [][]byte{nonce}
	assetEntry := NewEntry(oprChainID, entryExtIDs, bOpr)

	_, err = factom.CommitEntry(assetEntry, opr.EC)
	for i:=0;i<1000&&err!=nil;i++{
		time.Sleep(time.Duration(rand.Intn(50)+50)*time.Millisecond)
		_, err = factom.CommitEntry(assetEntry, opr.EC)
		if err == nil { break }
	}
	check(err)


//	assetEntry.Hash()


	factom.RevealEntry(assetEntry)
}

func GradeLastBlock(mstate *MinerState, opr *oprecord.OraclePriceRecord, dbht int64, miner *Mine) {

	var oprs []*oprecord.OraclePriceRecord

	oprChainID := mstate.GetOraclePriceRecordChain()

	var eb *factom.EBlock

	dblk,err := factom.GetDBlockByHeight(dbht)
	for i:=0; i < 15 && err != nil;i++{
		time.Sleep(100*time.Millisecond)
		dblk,err = factom.GetDBlockByHeight(dbht)
		if err == nil {break}
	}
	if err != nil {
		panic("Could not get the highest saved directory block.  "+err.Error())
	}

	dbentries := dblk.DBlock["dbentries"].([]interface{})
	for _,idbentry := range dbentries {
		dbentry := idbentry.(map[string]interface{})
		for key, v := range dbentry {
			if key == oprChainID {
				eb, err = factom.GetEBlock(v.(string))
			}
		}
	}

	// Get the last DirectoryBlock Merkle Root
	ebMR, err := factom.GetChainHead(oprChainID)
	for i:=0;i<1000 && err == nil; i++ {
		time.Sleep(time.Duration(rand.Intn(50)+50)*time.Millisecond)
		ebMR, err = factom.GetChainHead(oprChainID)
		if err == nil {break}
	}
	check(err)
	// Get the last DirectoryBlock
	eb, err = factom.GetEBlock(ebMR)
	for i:=0; i < 15 && (err != nil || eb == nil );i++ {
		eb, err = factom.GetEBlock(ebMR)
		if err == nil {
			break
		}
		time.Sleep(100*time.Millisecond)
	}
	if err != nil || eb == nil {
		fmt.Println(err)
		fmt.Printf("%s\n", ebMR)
	}

	for i, ebentry := range eb.EntryList {
		entry, err := factom.GetEntry(ebentry.EntryHash)
		if err != nil {
			fmt.Println(i, "Error Entry Nil")
			continue
		}
		if len(entry.ExtIDs) != 1 {
			fmt.Println(i, "Error ExtIDs not 1")
			continue
		}
		newOpr := new(oprecord.OraclePriceRecord)
		err = newOpr.UnmarshalBinary(entry.Content)
		if err != nil {
			fmt.Println(i, "Error Unmarshalling OPR")
			continue
		}

		// Compute the entry hash on the OPR record
		newOpr.GetEntry(mstate.GetOraclePriceRecordChain())

		rec := append([]byte{}, entry.ExtIDs[0]...)
		oprh := miner.HashFunction(entry.Content)
		rec = append(rec, oprh...)


		copy(newOpr.Nonce[:], entry.ExtIDs[0])
		copy(newOpr.OPRHash[:], oprh)

		if newOpr.ComputeDifficulty() == 0 {
			fmt.Println(i, "Error Difficulty is zero!")
			continue
		}

		oprs = append(oprs, newOpr)

	}

	tobepaid, oprlist := oprecord.GradeBlock(oprs)
	_, _ = tobepaid, oprlist
	if len(tobepaid) > 0 {
		copy(opr.WinningPreviousOPR[:], tobepaid[0].GetEntry(mstate.GetOraclePriceRecordChain()).Hash())

		//h := tobepaid[0].FactomDigitalID[:6]

			//fmt.Printf("OPRs %3d tobepaid %3d winner %x\n", len(oprs), len(tobepaid), h)

			if mstate.MinerNumber == 1 {
				for _, op := range tobepaid {
					fmt.Println(op.ShortString())
				}
		}
		if mstate.MinerNumber == 1 {
			fmt.Println(tobepaid[0].String())
		}
	}
}

func NewEntry(chainID string, extIDs [][]byte, content []byte) *factom.Entry {
	e := factom.Entry{ChainID: chainID, ExtIDs: extIDs, Content: content}
	return &e
}

func NewEntryStr(chainID string, extIDs []string, content string) *factom.Entry {
	var b [][]byte
	for _, str := range extIDs {
		b = append(b, []byte(str))
	}
	e := factom.Entry{ChainID: chainID, ExtIDs: b, Content: []byte(content)}
	return &e
}

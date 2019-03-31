package OracleMiner

import (
	"encoding/hex"
	"fmt"
	"github.com/FactomProject/FactomCode/common"
	"github.com/FactomProject/factom"
	"github.com/pegnet/LXR256"
	"github.com/pegnet/OracleRecord"
)

func InitNetwork(mstate *MinerState, minerNumber int, opr *oprecord.OraclePriceRecord) {

	PegNetChain := mstate.GetProtocolChain()
	BPegNetChainID,err := hex.DecodeString(PegNetChain)
	if err != nil {
		panic("Could not decode the protocol chain:"+err.Error())
	}
	opr.SetChainID(BPegNetChainID)

	bOprChainID,err := hex.DecodeString(mstate.GetProtocolChain())
	if err != nil {panic("No OPR Chain found in config file")}

	did := mstate.GetIdentityChainID()
	BFactomDigitalID, _ := hex.DecodeString(did)

	fmt.Println("ChainID: " + did)
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

	// Check that we have an asset entry
	entries, err := factom.GetAllChainEntries(chainid)
	check(err)
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
	if bal > 1000 {
		return
	}


	factom.DeleteTransaction("fundec")
	rate, err := factom.GetRate()
	check(err)

	_, err = factom.NewTransaction("fundec")
	check(err)

	fct2ec := uint64(1000) * rate // Buy so many FCT (shift left 8 decimal digits to create a fixed point number)
	_, err = factom.AddTransactionInput("fundec", fctadr, fct2ec)
	check(err)

	factom.AddTransactionECOutput("fundec", ecadr, fct2ec)
	_, err = factom.AddTransactionFee("fundec", fctadr)
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

	// Create the first entry for the PegNetChain
	sPegNetChainID := mstate.GetProtocolChain()
	sPegNetChainExtIDs := mstate.GetProtocolChainExtIDs()
	e := NewEntryStr(sPegNetChainID, sPegNetChainExtIDs, "")

	pegNetChain := factom.NewChain(e)

	txid, err := factom.CommitChain(pegNetChain, ec_adr)
	check(err)
	fmt.Println("Created network chain ", txid)
	factom.RevealChain(pegNetChain)
}

// The Pegged Network has a defining chain.  This function builds and returns the expected defining chain
// for the network.
func AddAssetEntry(mstate *MinerState) {
	fmt.Println("Adding AssetEntry")
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

	txid, err := factom.CommitEntry(assetEntry, ec_adr)
	check(err)
	fmt.Println("Created network chain ", txid)
	factom.RevealEntry(assetEntry)
}

// The Pegged Network has a defining chain.  This function builds and returns the expected defining chain
// for the network.
func CreateOPRChain(mstate *MinerState) {
	fmt.Println("Creating the Oracle Price Record chain")
	// Create an entry credit address
	ec_adr, err := factom.FetchECAddress(mstate.GetECAddress())
	// Create the first entry for the OPR Chain
	oprChainID := mstate.GetOraclePriceRecordChain()
    oprExtIDs := mstate.GetOraclePriceRecordExtIDs()

	e := NewEntryStr(oprChainID, oprExtIDs, "")

	oprChain := factom.NewChain(e)

	txid, err := factom.CommitChain(oprChain, ec_adr)
	check(err)
	fmt.Println("Created Oracle Price Record chain ", txid)
	factom.RevealChain(oprChain)
}

// The Pegged Network has a defining chain.  This function builds and returns the expected defining chain
// for the network.
func AddOpr(mstate *MinerState,  nonce []byte) {
	fmt.Println("Adding OPR Record")
	opr := mstate.OPR
	// Create the OPR Entry
	// Create the first entry for the OPR Chain
	oprChainID := mstate.GetOraclePriceRecordChain()
	bOpr, err := opr.MarshalBinary()
	check(err)

	entryExtIDs := [][]byte{nonce}
	assetEntry := NewEntry(oprChainID, entryExtIDs, bOpr)

	_, err = factom.CommitEntry(assetEntry, opr.EC)
	check(err)
	factom.RevealEntry(assetEntry)
}

func GradeLastBlock(mstate *MinerState, opr *oprecord.OraclePriceRecord, dbht int64, miner *Mine) {

	var oprs []*oprecord.OraclePriceRecord

	oprChainID := mstate.GetProtocolChain()
	// Get the last DirectoryBlock Merkle Root
	ebMR, err := factom.GetChainHead(oprChainID)
	check(err)
	// Get the last DirectoryBlock
	eb, err := factom.GetEBlock(ebMR)
	if err != nil {
		fmt.Println(err)
		fmt.Printf("%s\n", ebMR)
	}

	for _, ebentry := range eb.EntryList {
		entry, err := factom.GetEntry(ebentry.EntryHash)
		if err != nil {
			continue
		}
		if len(entry.ExtIDs) != 1 {
			continue
		}
		newOpr := new(oprecord.OraclePriceRecord)
		err = newOpr.UnmarshalBinary(entry.Content)
		if err != nil {
			continue
		}
		rec := append([]byte{}, entry.ExtIDs[0]...)
		oprh := miner.HashFunction(entry.Content)
		rec = append(rec, oprh...)
		h := miner.HashFunction(rec)
		diff := lxr.Difficulty(h)
		if diff == 0 {
			continue
		}
		copy(newOpr.Nonce[:], entry.ExtIDs[0])
		copy(newOpr.OPRHash[:], oprh)
		newOpr.Difficulty = diff
		oprs = append(oprs, newOpr)
	}

	tobepaid, oprlist := oprecord.GradeBlock(oprs)
	_, _ = tobepaid, oprlist
	if len(tobepaid) > 0 {
		copy(opr.WinningPreviousOPR[:], tobepaid[0].OPRHash[:])
	}
	if mstate.MinerNumber < 2 {
		for _, op := range oprlist {
			fmt.Println(op.String())
		}
	}
	h := []byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255}
	if len(tobepaid) > 0 {
		h = tobepaid[0].OPRHash[:]
	}
	fmt.Printf("Miner %3d oprs %3d tobepaid %3d winner %x\n", mstate.MinerNumber, len(oprs), len(tobepaid), h)
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

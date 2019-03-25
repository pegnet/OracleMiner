package OracleMiner

import (
	"encoding/hex"
	"fmt"
	"github.com/FactomProject/FactomCode/common"
	"github.com/FactomProject/factom"
	"github.com/pegnet/LXR256"
	"github.com/pegnet/OracleRecord"
	"io/ioutil"
	"strings"
)

func InitNetwork(minerNumber int, opr *oprecord.OraclePriceRecord) (NetworkChainID []byte) {
	sPegNetChainID := []string{"PegNet", "TestNet"}
	BPegNetChainID := factom.ComputeChainIDFromStrings(sPegNetChainID)
	opr.SetChainID(BPegNetChainID)
	NetworkChainID = append(NetworkChainID[:0], BPegNetChainID...)

	sOprExtIDs := []string{"Oracle Price Records", "TestNet"}
	bOprChainID := factom.ComputeChainIDFromStrings(sOprExtIDs)

	sFactomDID := []string{"prototype", "miner"}
	BFactomDigitalID := factom.ComputeChainIDFromStrings(sFactomDID)
	opr.SetFactomDigitalID(BFactomDigitalID)

	opr.SetVersionEntryHash(common.Sha([]byte("an entry")).Bytes())

	ecAdr, err := factom.FetchECAddress("EC3TsJHUs8bzbbVnratBafub6toRYdgzgbR7kWwCW4tqbmyySRmg")
	if err != nil {
		fmt.Println("Failed to initialize ", minerNumber)
		fmt.Println(err.Error())
		return
	}
	opr.EC = ecAdr

	FundWallet()

	// First check if the network has been initialized.  If it hasn't, then create all
	// the initial structures.  This is only needed while testing.
	chainid := hex.EncodeToString(opr.ChainID[:])
	if !factom.ChainExists(chainid) {
		CreatePegNetChain(opr)
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
		AddAssetEntry(opr)
	}()

	// Check for and create the Oracle Price Records Chain
	chainid = hex.EncodeToString(bOprChainID)
	if !factom.ChainExists(chainid) {
		CreateOPRChain(opr)
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
func FundWallet() (err error) {
	// Get our EC address
	ecadr := "EC3TsJHUs8bzbbVnratBafub6toRYdgzgbR7kWwCW4tqbmyySRmg"

	// Check and see if we have at least 1000 entry credits
	bal, err := factom.GetECBalance(ecadr)
	check(err)
	// If we have the 1000 entry credits, then we are happy clams, and can move on!
	if bal > 1000 {
		fmt.Println("We have ", bal, " entry credits!")
		return
	}
	// Ah, we need the EC!
	dat, err := ioutil.ReadFile("fct.dat")
	check(err)
	fctadr := strings.TrimSpace(string(dat))

	factom.DeleteTransaction("fundec")
	rate, err := factom.GetRate()
	check(err)

	_, err = factom.NewTransaction("fundec")
	check(err)

	fct2ec := uint64(100) * 100000000 // Buy so many FCT (shift left 8 decimal digits to create a fixed point number)
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
func CreatePegNetChain(opr *oprecord.OraclePriceRecord) {
	fmt.Println("Creating the PegNetChain")
	// Create an entry credit address
	ec_adr, err := factom.FetchECAddress("EC3TsJHUs8bzbbVnratBafub6toRYdgzgbR7kWwCW4tqbmyySRmg")

	// Create the first entry for the PegNetChain
	sPegNetChainID := [][]byte{[]byte("PegNet"), []byte("TestNet")}
	PegNetChainID := hex.EncodeToString(opr.ChainID[:])
	SPegNetChainID := hex.EncodeToString(factom.ComputeChainIDFromFields(sPegNetChainID))
	_ = SPegNetChainID
	e := NewEntry(PegNetChainID, sPegNetChainID, []byte{})

	pegNetChain := factom.NewChain(e)

	txid, err := factom.CommitChain(pegNetChain, ec_adr)
	check(err)
	fmt.Println("Created network chain ", txid)
	factom.RevealChain(pegNetChain)
}

// The Pegged Network has a defining chain.  This function builds and returns the expected defining chain
// for the network.
func AddAssetEntry(opr *oprecord.OraclePriceRecord) {
	fmt.Println("Adding AssetEntry")
	// Create an entry credit address
	ec_adr, err := factom.FetchECAddress("EC3TsJHUs8bzbbVnratBafub6toRYdgzgbR7kWwCW4tqbmyySRmg")

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
	PegNetChainID := hex.EncodeToString(opr.ChainID[:])

	assetEntry := NewEntryStr(PegNetChainID, assets, "")

	txid, err := factom.CommitEntry(assetEntry, ec_adr)
	check(err)
	fmt.Println("Created network chain ", txid)
	factom.RevealEntry(assetEntry)
}

// The Pegged Network has a defining chain.  This function builds and returns the expected defining chain
// for the network.
func CreateOPRChain(opr *oprecord.OraclePriceRecord) {
	fmt.Println("Creating the Oracle Price Record chain")
	// Create an entry credit address
	ec_adr, err := factom.FetchECAddress("EC3TsJHUs8bzbbVnratBafub6toRYdgzgbR7kWwCW4tqbmyySRmg")

	// Create the first entry for the OPR Chain
	oprExtIDs := []string{"Oracle Price Records", "TestNet"}
	oprChainID := hex.EncodeToString(factom.ComputeChainIDFromStrings(oprExtIDs))

	e := NewEntryStr(oprChainID, oprExtIDs, "")

	oprChain := factom.NewChain(e)

	txid, err := factom.CommitChain(oprChain, ec_adr)
	check(err)
	fmt.Println("Created Oracle Price Record chain ", txid)
	factom.RevealChain(oprChain)
}

// The Pegged Network has a defining chain.  This function builds and returns the expected defining chain
// for the network.
func AddOpr(opr *oprecord.OraclePriceRecord, nonce []byte) {
	fmt.Println("Adding OPR Record")

	// Create the OPR Entry
	// Create the first entry for the OPR Chain
	oprExtIDs := []string{"Oracle Price Records", "TestNet"}
	oprChainID := hex.EncodeToString(factom.ComputeChainIDFromStrings(oprExtIDs))
	bOpr, err := opr.MarshalBinary()
	check(err)

	entryExtIDs := [][]byte{nonce}
	assetEntry := NewEntry(oprChainID, entryExtIDs, bOpr)

	txid, err := factom.CommitEntry(assetEntry, opr.EC)
	check(err)
	fmt.Println("Wrote OPR Record", txid)
	factom.RevealEntry(assetEntry)
}

func GradeLastBlock(opr *oprecord.OraclePriceRecord, dbht int64, miner *Mine) {

	var oprs []*oprecord.OraclePriceRecord

	oprExtIDs := []string{"Oracle Price Records", "TestNet"}
	oprChainID := hex.EncodeToString(factom.ComputeChainIDFromStrings(oprExtIDs))
	// Get the last DirectoryBlock Merkle Root
	ebMR, err := factom.GetChainHead(oprChainID)
	check(err)
	// Get the last DirectoryBlock
	eb, err := factom.GetEBlock(ebMR)
	check(err)

	fmt.Printf("Got chain head at height %d\n", eb.Header.DBHeight)

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
		fmt.Printf("Review entry hash %s nonce %x oprh %x hash %x\n", ebentry.EntryHash, entry.ExtIDs[0], oprh, h)
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

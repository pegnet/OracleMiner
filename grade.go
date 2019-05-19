package OracleMiner

import (
	"fmt"
	"github.com/FactomProject/factom"
	"github.com/pegnet/OracleRecord"
)

func GradeLastBlock(mstate *MinerState, opr *oprecord.OraclePriceRecord, dbht int64, miner *Mine) {

	var oprs []*oprecord.OraclePriceRecord

	oprChainID := mstate.GetOraclePriceRecordChain()

	var eb *factom.EBlock

	// Get the last DirectoryBlock Merkle Root
	ebMR, err := factom.GetChainHead(oprChainID)
	check(err)
	// Get the last DirectoryBlock
	eb, err = factom.GetEBlock(ebMR)
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
		if newOpr.Dbht != int32(dbht) {
			//continue
		}
		newOpr.Entry = entry
		copy(newOpr.Nonce[:], entry.ExtIDs[0])

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
			for i, op := range tobepaid {
				fmt.Printf("%3d %s\n", i+1, op.ShortString())
			}
			fmt.Println("==================")
			for i, op := range oprlist {
				fmt.Printf("%3d %s\n", i+1, op.ShortString())
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

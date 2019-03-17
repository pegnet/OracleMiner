package OracleMiner

import (
	"encoding/hex"
	"github.com/FactomProject/factom"
	"github.com/pegnet/OracleRecord"
	"github.com/FactomProject/FactomCode/common"
)



func GetChainID(fields []string) []byte {
	var binary [][]byte
	for _, str := range fields {
		b, err := hex.DecodeString(str)
		if err != nil {
			panic(err)
		}
		binary = append(binary, b)
	}
	return factom.ComputeChainIDFromFields(binary)
}

func InitNetwork(opr *oprecord.OraclePriceRecord) {
	sPegNetChainID := []string{"PegNet", "TestNet"}
	BPegNetChainID := GetChainID(sPegNetChainID)
	opr.SetChainID(BPegNetChainID)

	sFactomDID := []string{"prototype", "miner"}
	BFactomDigitalID := GetChainID(sFactomDID)
	opr.SetFactomDigitalID(BFactomDigitalID)

	opr.SetVersionEntryHash(common.Sha([]byte("an entry")).Bytes())
}



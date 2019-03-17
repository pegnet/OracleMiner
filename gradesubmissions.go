package OracleMiner

import (
	"github.com/FactomProject/factom"
	"github.com/pegnet/OracleRecord"
	"github.com/FactomProject/FactomCode/common"
)


// GetChainIDFromStrings
// Take a set of strings, and compute the chainID.  If you have binary fields, you
// can call factom.ComputeChainIDFromFields directly, which takes [][]byte
func GetChainIDFromStrings(fields []string) []byte {
	var binary [][]byte
	for _, str := range fields {
		binary = append(binary, []byte(str))
	}
	return factom.ComputeChainIDFromFields(binary)
}

func InitNetwork(opr *oprecord.OraclePriceRecord) {
	sPegNetChainID := []string{"PegNet", "TestNet"}
	BPegNetChainID := GetChainIDFromStrings(sPegNetChainID)
	opr.SetChainID(BPegNetChainID)

	sFactomDID := []string{"prototype", "miner"}
	BFactomDigitalID := GetChainIDFromStrings(sFactomDID)
	opr.SetFactomDigitalID(BFactomDigitalID)

	opr.SetVersionEntryHash(common.Sha([]byte("an entry")).Bytes())
}



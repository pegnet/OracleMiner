package OracleMiner

import (
	"encoding/hex"
	"fmt"
	"github.com/FactomProject/factom"
	"github.com/pegnet/OracleRecord"
	"github.com/zpatrick/go-config"
	"os/user"
	"strings"
)

const protocolname = "zoints"

type MinerState struct {
	MinerNumber int                        // If running multiple miners, this is the number
	Monitor     *FactomdMonitor			   // The facility that tracks what blocks we are processing
	ConfigDir   string                     // Must end with a /
	OPR         oprecord.OraclePriceRecord // The price record we mine against
	Config      *config.Config             // Configuration file with data for mining and Oracle data
}

func (m *MinerState) LoadConfig() {
	u, err := user.Current()
	if err != nil {
		panic(err)
	}
	userPath := u.HomeDir
	configfile := fmt.Sprintf("%s/.%s/miner%03d/config.ini", userPath, protocolname, m.MinerNumber)
	iniFile := config.NewINIFile(configfile)
	m.Config = config.NewConfig([]config.Provider{iniFile})
	_, err = m.Config.String("Miner.Protocol")
	if err != nil {
		configfile = fmt.Sprintf("%s/.%s/defaultconfig.ini", userPath, protocolname)
		iniFile := config.NewINIFile(configfile)
		m.Config = config.NewConfig([]config.Provider{iniFile})
		_, err = m.Config.String("Miner.Protocol")
		if err != nil {
			panic("Failed to open the config file for this miner, and couldn't load the default file either")
		}
	}
}

func (m *MinerState) GetECAddress() string {
	if str, err := m.Config.String("Miner.ECAddress"); err != nil {
		panic("No Entry Credit address in Config" + err.Error())
	} else {
		return str
	}
}

func (m *MinerState) GetFCTAddress() string {
	if str, err := m.Config.String("Miner.FCTAddress"); err != nil {
		panic("No FCT address in Config" + err.Error())
	} else {
		return str
	}
}

func (m *MinerState) GetProtocolChainExtIDs() []string {
	protocol, err1 := m.Config.String("Miner.Protocol")
	network, err2 := m.Config.String("Miner.Network")
	if err1 != nil || err2 != nil {
		panic("Missing either the Protocol or Network entries in the Config file")
	}
	extIDs := []string{protocol, network}
	return extIDs
}

// GetProtocolChain
// Get the chain for the protocol.  Versions of the protocol (when updates are deployed) will be recorded
// in this chain.
func (m *MinerState) GetProtocolChain() string {
	chainid := factom.ComputeChainIDFromStrings(m.GetProtocolChainExtIDs())
	return hex.EncodeToString(chainid)
}

func (m *MinerState) GetOraclePriceRecordExtIDs() []string {
	protocol, err1 := m.Config.String("Miner.Protocol")
	network, err2 := m.Config.String("Miner.Network")
	if err1 != nil || err2 != nil {
		panic("Missing either the Protocol or Network entries in the Config file")
	}
	sOprExtIDs := []string{protocol, network, "OPR"}
	return sOprExtIDs
}

// GetOraclePriceRecordChain
// Returns the chainID for all the Oracle Price Records.  Note that this ID is the same as the
// ProtocolChain + the field "OPR"
func (m *MinerState) GetOraclePriceRecordChain() string {
	chainid := factom.ComputeChainIDFromStrings(m.GetOraclePriceRecordExtIDs())
	return hex.EncodeToString(chainid)
}

// GetIdentityChainID
// Returns a pointer to a string for the chainID.  Takes the raw chain
// if specified, but if not, returns the chainID computed from the fields.
// If no chainID is specified in the config file, a nil is returned.
func (m *MinerState) GetIdentityChainID() string {
	chainID, err := m.Config.String("Miner.IdentityChain")
	if err != nil || len(chainID) == 0 {
		fieldscomma, err := m.Config.String("Miner.IdentityChainFields")
		fields := strings.Split(fieldscomma, ",")
		if err != nil || len(fields) == 0 {
			panic("Could not find the Identity Chain ID or Identity Chain Fields for the miner")
		}
		if len(fields) == 1 && string(fields[0]) == "prototype" {
			fields = append(fields, fmt.Sprintf("miner%03d", m.MinerNumber))
		}
		bchainID := factom.ComputeChainIDFromStrings(fields)
		return hex.EncodeToString(bchainID)
	}
	return chainID
}

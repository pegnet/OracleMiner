package OMconfig

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestConfig(t *testing.T) {
	configFileStuff := "" +
		"[Miner]\n" +
		"   Protocol=PegNet\n" +
		"   Network=TestNet\n" +
		"   ECAddress=EC3TsJHUs8bzbbVnratBafub6toRYdgzgbR7kWwCW4tqbmyySRmg\n" +
		"   IdentityChain=45b713101889a561df4028b3197459d2fca9783745d996ae090f672c8387914d\n" +
		"   IdentityChainFields=prototype,miner"
	dir := os.TempDir() + "/"
	ioutil.WriteFile(dir+"config.ini", []byte(configFileStuff), 0777)
	c := loadConfig(dir)
	Protocol, err1 := c.String("Miner.Protocol")
	Network, err2 := c.String("Miner.Network")
	ECAddress, err3 := c.String("Miner.ECAddress")
	IdentityChain, err4 := c.String("Miner.IdentityChain")
	Fields, err5 := c.String("Miner.IdentityChainFields")
	chainID := GetChainIDFromFields(Fields)

	if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil {
		t.Error("Got an error calling c.String()")
	}
	if Protocol != "PegNet" ||
		Network != "TestNet" ||
		ECAddress != "EC3TsJHUs8bzbbVnratBafub6toRYdgzgbR7kWwCW4tqbmyySRmg" ||
		IdentityChain != "45b713101889a561df4028b3197459d2fca9783745d996ae090f672c8387914d" ||
		chainID != IdentityChain {
		t.Error("Didn't load the values")
	}

}

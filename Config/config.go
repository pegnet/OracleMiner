package OMconfig

import (
	"github.com/zpatrick/go-config"
)

// loadConfig
// Filepath needs to be of the form "~/.oracleminer/miner01/"
func loadConfig(filepath string) *config.Config {
	iniFile := config.NewINIFile(filepath + "config.ini")
	c := config.NewConfig([]config.Provider{iniFile})
	if err := c.Load(); err != nil {
		panic("Failed to load Configuration: " + err.Error())
	}
	return c
}

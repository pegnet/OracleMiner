# RunMiner
This code creates a fake OPR record, and spins up a specified number of miner 
processes to mine that OPR record. 

The code as written must be modified to change the number of mining processes, 
and/or the block times (i.e. assume 10 second blocks, 1 minute blocks, etc.)

With every block, the hash rate over the block, the difficulty found, and other
statistics are printed.

TODO:

Write unit tests, parametrise the block time and miner processes.

# RunMiner2

This code assumes a running factomd node.  Defaults to localhost. 

This code creates a fake OPR record, and spins up a specified number of miner 
processes to mine that OPR record. 

We pole the factomd node for the start of the block, then we mine until minute 8.

With every block, the hash rate over the block, the difficulty found, and other
statistics are printed.

TODO:

Write unit tests, parametrise the block time and miner processes.

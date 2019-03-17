# FactomdMining

This is the code that actually runs the mining for the Pegged Network

1. Run the factomd daemon with its default configuration
2. Run the factom-walletd daemon with its default configuration
3. Make sure you have your Api Access Key in the apikey.dat file in the directory
   that you launch PeggedNetworkMining
4. Add your Entry Credit address in your ecaddress.dat file in the directory  
   that you launch PeggedNetworkMining
5. Make sure your Entry Credit Address is funded
6. Launch PeggednetworkMining


This code assumes a running factomd node.  Defaults to localhost:8088 but could be pointed 
at any running factomd node.

Code gets a Factom Record

We pole the factomd node for the start of the block, then we mine until minute 8.

With every block, the hash rate over the block, the difficulty found, and other
statistics are printed.

TODO:

Write unit tests, parametrise the block time and miner processes.

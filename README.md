
# ABCI Proxy
- launch golang code
--> tendermint docker (0.10.0) use it as it is.
--> ABCI proxy (we need to work on it)
--> Counter Example Application (https://github.com/tendermint/abci/tree/master/example/counter)
# 1st Goal
--> make sure the 3 instances talk to each other.

# 2nd Goal
- run the 3 dockers on 2 nodes (2 first nodes are validators, 1 others just observer)

when the block 1000 arrives, then ABCi proxy change the 1 observers to validators

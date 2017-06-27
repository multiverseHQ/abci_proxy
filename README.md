
# ABCI Proxy

- launch golang code

--> tendermint docker (0.10.0) use it as it is.

--> ABCI proxy (in this github Repo.)

--> Counter Example Application (https://github.com/multiverseHQ/abci_sample)

# 1st Goal

--> make sure the 3 instances talk to each other.

# 2nd Goal
- run the 3 instances on 5 nodes (4 first nodes are validators, 1 others just observer)

when the block 1000 arrives, then ABCi proxy change the 1 observer to 1 validator

# 1st Goal
- create 3 docker container 
--> tendermint docker (0.10.0)
--> ABCI proxy
--> Counter Example Application (https://github.com/tendermint/abci/tree/master/example/counter)

- run the 3 dockers on 5 nodes (3 first nodes are validators, 2 others just observer)

when the block 1000 arrives, then ABCi proxy change the 2 observers to validators

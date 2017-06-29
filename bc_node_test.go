package abciproxy

import "fmt"

type BCNodeID int

// A BCNode consist of a tendermint node, its application and the
// proxy running between those This object is here to maintain sane
// configuration and overall start/stop function for testing purpose
// on a single computer
type BCNode struct {
	ID BCNodeID
}

const MaxBCNodeID BCNodeID = 99

const ProxyAppPortStart int = 50000
const RPCPortStart int = 50100
const P2PPortStart int = 50200
const AppPortStart int = 50300

func NewBCNode(ID BCNodeID, genesisPath string, homeDir string) (*BCNode, error) {
	if ID > MaxBCNodeID {
		return nil, fmt.Errorf("Maximum  BCNodeID for test is %d, got %d", MaxBCNodeID, ID)
	}

	return nil, NotYetImplemented()
}

func (*BCNode) Start() error {
	return NotYetImplemented()
}

func (*BCNode) Stop() error {
	return NotYetImplemented()
}

func (n *BCNode) ProxyAppPort() int {
	return ProxyAppPortStart + int(n.ID)
}

func (n *BCNode) RPCPort() int {
	return RPCPortStart + int(n.ID)
}

func (n *BCNode) P2PPort() int {
	return RPCPortStart + int(n.ID)
}

func (n *BCNode) AppPort() int {
	return AppPortStart + int(n.ID)
}

package abciproxy

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type UseCaseSuite struct {
	testHome    string
	genesisPath string
	nodes       []*BCNode
}

var _ = Suite(&UseCaseSuite{})

// SetUpGenesis setup a new genesis file for working on a new blockchain. Validators are the ID of the validators with a weight of one, and observers are the one with a weight of 1
func (s *UseCaseSuite) SetUpGenesis(validators, observers []*BCNode) error {

	//TODO: create a new blockchain, via the correct genesis file, with the right information
	return NotYetImplemented()
}

func (s *UseCaseSuite) SetUpSuite(c *C) {
	var err error
	//we create a new tendermint structure for the test

	s.testHome, err = ioutil.TempDir("", "abci_proxy_test")
	c.Assert(err, IsNil, Commentf("Cannot create tmpdir: %s", err))

	s.genesisPath = filepath.Join(s.testHome, "genesis.json")

	//TODO: create the 4 required node
	s.nodes = nil
	for i := 0; i < 4; i++ {
		//create a new home directory for the node
		nDir, err := ioutil.TempDir(s.testHome, "blockchain_node")
		c.Assert(err, IsNil)
		n, err := NewBCNode(BCNodeID(i), s.genesisPath, nDir)
		c.Assert(err, IsNil, Commentf("Could not create node %d: %s", i, err))
		s.nodes = append(s.nodes, n)
	}

}

func (s *UseCaseSuite) TearDownSuite(c *C) {
	err := os.RemoveAll(s.testHome)
	c.Check(err, IsNil, Commentf("cannot remove temporary directory: %s", err))
}

func (s *UseCaseSuite) SetUpTest(c *C) {
	//create a new blockchain to work with
	err := s.SetUpGenesis(s.nodes[0:3], s.nodes[3:])
	c.Assert(err, IsNil, Commentf("Could not setup genesis file:%s", err))

	for _, n := range s.nodes {
		err = n.Start()
		c.Assert(err, IsNil)
	}
}

func (s *UseCaseSuite) TearDownTest(c *C) {
	for _, n := range s.nodes {
		err := n.Stop()
		c.Assert(err, IsNil)
	}
}

func (s *UseCaseSuite) TestChangeObserverToValidator(c *C) {
	//the 4 chain are started

	//specify that when height = 100, we should change observer node 3 to validators

}

func (s *UseCaseSuite) TestAddValidatorOnTheFly(c *C) {
	//the 4 chain are started

}

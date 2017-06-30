package abciproxy

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type UseCaseSuite struct {
	testHome     string
	genesisPath  string
	nodes        []*BCNode
	startedNodes []*BCNode
}

var _ = Suite(&UseCaseSuite{})

// SetUpGenesis setup a new genesis file for working on a new blockchain. Validators are the ID of the validators with a weight of one, and observers are the one with a weight of 1
func (s *UseCaseSuite) SetUpGenesis(validators, observers []*BCNode) error {

	//TODO: create a new blockchain, via the correct genesis file, with the right information
	func() {
		log.Printf("W A R N I N G\nBlockchain are independant and does not share the same genesis file, please implement %s", CallerName())
	}()
	return nil
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
		n, err := NewBCNode(BCNodeID(i), s.genesisPath, s.testHome)
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
}

func (s *UseCaseSuite) StartNodes(c *C) {
	log.Printf("Starting up blockchain nodes")
	for _, n := range s.nodes {
		err := n.Start()
		c.Assert(err, IsNil)
		s.startedNodes = append(s.startedNodes, n)
	}
}

func (s *UseCaseSuite) TearDownTest(c *C) {
	log.Printf("Closing up started blockchain nodes")
	for _, n := range s.startedNodes {
		err := n.Stop()
		c.Assert(err, IsNil)
	}
}

func (s *UseCaseSuite) TestCanRunSimpleCounterInOneNode(c *C) {

	//expect 10 calls
	s.nodes[0].testApplication.EndBlockCalls.ExpectCall(10)

	err := s.nodes[0].Start()
	c.Assert(err, IsNil)

	s.nodes[0].testApplication.EndBlockCalls.WaitForExpected()

	//for debugging purposes
	log.Printf("\n---tendermint output---\n%s\n---proxy output---\n%s\n---app output---\n%s",
		s.nodes[0].tmNodeOutput.String(),
		s.nodes[0].proxyOutput.String(),
		s.nodes[0].appOutput.String())
	//specify that when height = 100, we should change observer node 3 to validators

	//manually closing up the node
	err = s.nodes[0].Stop()
	c.Assert(err, IsNil)
}

// func (s *UseCaseSuite) TestAddValidatorOnTheFly(c *C) {
// 	// the 4 chain are started

// }

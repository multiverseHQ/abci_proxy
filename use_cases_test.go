package abciproxy

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/tendermint/tendermint/types"

	. "gopkg.in/check.v1"
)

// Number of node to instantiate for tests
const TotalTestNode = 5

func Test(t *testing.T) { TestingT(t) }

type UseCaseSuite struct {
	testHome      string
	genesisPath   string
	nodes         []*BCNode
	startedNodes  []*BCNode
	genesisFiles  map[BCNodeID]*types.GenesisDoc
	genesisNumber int
}

var _ = Suite(&UseCaseSuite{})

// SetUpGenesis setup a new genesis file for working on a new
// blockchain. Validators are the ID of the validators with a weight
// of one, and observers are the one with a weight of 1
func (s *UseCaseSuite) SetUpGenesis(validators, observers []*BCNode) error {
	s.genesisNumber++
	firstGenesis := s.genesisFiles[s.nodes[0].ID]

	//merge all of them in a single nice genesis file
	newGenesis := types.GenesisDoc{
		GenesisTime: firstGenesis.GenesisTime,
		ChainID:     firstGenesis.ChainID + fmt.Sprintf("-test-%d", s.genesisNumber),
		AppHash:     firstGenesis.AppHash,
	}

	for _, n := range validators {
		validator := s.genesisFiles[n.ID].Validators[0]
		newGenesis.Validators = append(newGenesis.Validators, validator)
	}

	for _, n := range observers {
		validator := s.genesisFiles[n.ID].Validators[0]
		validator.Amount = 0
		newGenesis.Validators = append(newGenesis.Validators, validator)
	}

	return newGenesis.SaveAs(s.genesisPath)
}

// SetUpSuite creates the temporary structure for all nodes we requires
func (s *UseCaseSuite) SetUpSuite(c *C) {
	var err error

	// everything goes in a nice temporary directory
	s.testHome, err = ioutil.TempDir("", "abci_proxy_test")
	c.Assert(err, IsNil, Commentf("Cannot create tmpdir: %s", err))

	s.genesisPath = filepath.Join(s.testHome, "genesis.json")

	s.nodes = nil
	for i := 0; i < TotalTestNode; i++ {
		n, err := NewBCNode(BCNodeID(i), s.genesisPath, s.testHome)
		c.Assert(err, IsNil, Commentf("Could not create node %d: %s", i, err))
		s.nodes = append(s.nodes, n)
	}

	//gather all nodes genesis
	s.genesisFiles = make(map[BCNodeID]*types.GenesisDoc)
	for _, n := range s.nodes {
		gPath := filepath.Join(n.WorkingDir(), "genesis.json")
		gDoc, err := types.GenesisDocFromFile(gPath)
		c.Assert(err, IsNil)
		s.genesisFiles[n.ID] = gDoc

	}

	s.genesisNumber = 0

	//ugly way to specify the genesis, sorry
	configToml := `# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

proxy_app = "tcp://127.0.0.1:46658"
moniker = "anonymous"
fast_sync = true
db_backend = "leveldb"
log_level = "state:info,*:error"
genesis_file = "` + s.genesisPath + `"

[rpc]
laddr = "tcp://0.0.0.0:46657"

[p2p]
laddr = "tcp://0.0.0.0:46656"
seeds = ""
`
	for _, n := range s.nodes {
		configPath := filepath.Join(n.WorkingDir(), "config.toml")
		err := ioutil.WriteFile(configPath, []byte(configToml), 0644)
		c.Assert(err, IsNil)
	}

}

// TearDownSuite cleans all temporary data
func (s *UseCaseSuite) TearDownSuite(c *C) {
	err := os.RemoveAll(s.testHome)
	c.Check(err, IsNil, Commentf("cannot remove temporary directory: %s", err))
}

// SetUpTest creates a new single genesis file for each of the test
func (s *UseCaseSuite) SetUpTest(c *C) {
	//create a new blockchain to work with first 4 nodes validators and last observer
	err := s.SetUpGenesis(s.nodes[0:(TotalTestNode-1)], s.nodes[(TotalTestNode-1):])
	c.Assert(err, IsNil, Commentf("Could not setup genesis file:%s", err))
}

// StartNodes starts all the instantiated nodes and mark them to be
// stopped by TearDownTest. You have to call it manually in your test
func (s *UseCaseSuite) StartNodes(c *C) {
	log.Printf("Starting up blockchain nodes")
	for _, n := range s.nodes {
		err := n.Start(s.nodes)
		c.Assert(err, IsNil)
		s.startedNodes = append(s.startedNodes, n)
	}
}

// TearDownTest ensure that after every test, we stop all the node which have been started with StartNodes
func (s *UseCaseSuite) TearDownTest(c *C) {
	log.Printf("Closing up started blockchain nodes")
	for _, n := range s.startedNodes {
		err := n.Stop()
		c.Assert(err, IsNil)
	}
}

func (s *UseCaseSuite) TestCanRunSimpleCounterInAllNode(c *C) {
	// remarks: this test is for a single node (nodes[0]) we start and stop it manually

	//expect 10 calls
	for _, n := range s.nodes {
		n.testApplication.EndBlockCalls.ExpectCall(10)
	}

	s.StartNodes(c)

	for _, n := range s.nodes {
		n.testApplication.EndBlockCalls.WaitForExpected()
		//for debugging purposes
		log.Printf("\n---tendermint output [Node %d] ---\n%s\n---proxy output  [Node %d] ---\n%s\n---app output  [Node %d]---\n%s",
			n.ID, n.tmNodeOutput.String(),
			n.ID, n.proxyOutput.String(),
			n.ID, n.appOutput.String())
	}
	//specify that when height = 100, we should change observer node 3 to validators

}

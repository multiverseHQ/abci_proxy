package abciproxy

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	abcitypes "github.com/tendermint/abci/types"
	rpctypes "github.com/tendermint/tendermint/rpc/core/types"
	rpc "github.com/tendermint/tendermint/rpc/lib/client"
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
	s.genesisNumber = s.genesisNumber + 1

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

	os.Remove(s.genesisPath)

	return newGenesis.SaveAs(s.genesisPath)
}

// SetUpSuite creates the temporary structure for all nodes we requires
func (s *UseCaseSuite) SetUpSuite(c *C) {
	var err error

	s.genesisNumber = 0

	// everything goes in a nice temporary directory
	s.testHome, err = ioutil.TempDir("", "abci_proxy_test")
	c.Assert(err, IsNil, Commentf("Cannot create tmpdir: %s", err))

	s.genesisPath = filepath.Join(s.testHome, "genesis.json")

}

// TearDownSuite cleans all temporary data
func (s *UseCaseSuite) TearDownSuite(c *C) {
	err := os.RemoveAll(s.testHome)
	c.Check(err, IsNil, Commentf("cannot remove temporary directory: %s", err))
	os.Remove(s.testHome)
}

// SetUpTest creates a new single genesis file for each of the test
func (s *UseCaseSuite) SetUpTest(c *C) {
	s.nodes = nil
	for i := 0; i < TotalTestNode; i++ {
		n, err := NewBCNode(BCNodeID(i), s.testHome)
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
		err = os.RemoveAll(n.WorkingDir())
		c.Assert(err, IsNil)
	}

	s.startedNodes = make([]*BCNode, 0, len(s.nodes))
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

		// log.Printf("[Node %d Calls]\nInfo: %d\nCommit: %d\nSetoption: %d\nDeliverTx: %d\nCheckTx: %d\nQuery %d\nEndBlock: %d", n.ID,
		// 	len(n.testApplication.InfoCalls.Calls),
		// 	len(n.testApplication.CommitCalls.Calls),
		// 	len(n.testApplication.SetOptionCalls.Calls),
		// 	len(n.testApplication.DeliverTxCalls.Calls),
		// 	len(n.testApplication.CheckTxCalls.Calls),
		// 	len(n.testApplication.QueryCalls.Calls),
		// 	len(n.testApplication.EndBlockCalls.Calls),
		// )

		//ensure all node reached height 10
		endBlockCalls := n.testApplication.EndBlockCalls.Calls
		lastHeight := endBlockCalls[len(endBlockCalls)-1][0].(uint64)
		c.Check(lastHeight >= uint64(10), Equals, true, Commentf("node %d reached %d", n.ID, lastHeight))

	}

}

func (s *UseCaseSuite) TestCanMakeObserverAValidator(c *C) {

	//wait the block 10 on first node
	s.nodes[0].testApplication.EndBlockCalls.ExpectCall(5)

	observerKey := s.genesisFiles[TotalTestNode-1].Validators[0].PubKey

	s.StartNodes(c)

	validators := make([]*abcitypes.Validator, 0, len(s.nodes))

	//build the new list of Validator
	for _, g := range s.genesisFiles {
		c.Assert(len(g.Validators), Equals, 1)
		v := &abcitypes.Validator{
			PubKey: g.Validators[0].PubKey.Bytes(),
			Power:  10,
		}
		validators = append(validators, v)
	}

	s.nodes[0].testApplication.EndBlockCalls.WaitForExpected()
	for _, n := range s.nodes {
		n.proxy.ChangeValidators(validators, 10)
		n.testApplication.EndBlockCalls.ExpectCall(4)
	}

	//we check for the next 2 block, the validator does not change
	for _, n := range s.nodes {
		n.testApplication.EndBlockCalls.WaitForExpected()

		//we check that all the validator have the correct voting power, aka 10 except for the observer
		client := rpc.NewJSONRPCClient(fmt.Sprintf("http://localhost:%d", n.RPCPort()))

		res := new(rpctypes.ResultValidators)
		_, err := client.Call("validators", map[string]interface{}{}, res)
		if c.Check(err, IsNil) == false {
			continue
		}
		for _, v := range res.Validators {
			expectedVotingPower := int64(10)
			if v.PubKey == observerKey {
				expectedVotingPower = int64(0)
			}

			c.Check(v.VotingPower, Equals, expectedVotingPower)
		}

		n.testApplication.EndBlockCalls.ExpectCall(2)
	}

	//we check that after block 11 all validators are here
	for _, n := range s.nodes {
		n.testApplication.EndBlockCalls.WaitForExpected()

		//ensure all node reached height 11
		endBlockCalls := n.testApplication.EndBlockCalls.Calls
		lastHeight := endBlockCalls[len(endBlockCalls)-1][0].(uint64)
		c.Check(lastHeight >= uint64(11), Equals, true, Commentf("node %d reached %d", n.ID, lastHeight))

		//we check that all the validator have the correct voting power

		client := rpc.NewJSONRPCClient(fmt.Sprintf("http://localhost:%d", n.RPCPort()))

		res := new(rpctypes.ResultValidators)
		_, err := client.Call("validators", map[string]interface{}{}, res)
		if c.Check(err, IsNil) == false {
			continue
		}
		for _, v := range res.Validators {
			c.Check(v.VotingPower, Equals, int64(10))
		}
	}

}

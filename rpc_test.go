package abciproxy

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/tendermint/abci/types"
	"github.com/tendermint/tendermint/rpc/lib/client"
	tmtypes "github.com/tendermint/tendermint/types"

	. "gopkg.in/check.v1"
)

type RPCSuite struct {
	testHome    string
	node        *BCNode
	genesisFile *tmtypes.GenesisDoc

	cli *rpcclient.JSONRPCClient
}

var _ = Suite(&RPCSuite{})

func (s *RPCSuite) SetUpSuite(c *C) {
	var err error

	s.testHome, err = ioutil.TempDir("", "abci_proxy_test")
	c.Assert(err, IsNil, Commentf("Cannot create tempdir: %s", err))

	s.node, err = NewBCNode(0, s.testHome)
	c.Assert(err, IsNil, Commentf("Cannot create node: %s", err))

	err = s.node.Start(nil)
	c.Assert(err, IsNil)

	//only in this suite please, as we cannot stop it
	address := fmt.Sprintf("127.0.0.1:%d", s.node.RPCProxyPort())
	s.node.proxy.StartRPCServer("tcp://" + address)

	s.genesisFile, err = tmtypes.GenesisDocFromFile(filepath.Join(s.node.WorkingDir(), "genesis.json"))
	c.Assert(err, IsNil)

	s.cli = rpcclient.NewJSONRPCClient("http://" + address)
}

func (s *RPCSuite) TearDownSuite(c *C) {
	err := s.node.Stop()
	c.Assert(err, IsNil)

	err = os.RemoveAll(s.testHome)
	c.Assert(err, IsNil)

	os.Remove(s.testHome)
}

func (s *RPCSuite) TestCanFetchCurrentHeight(c *C) {
	//wait for at least one call
	s.node.testApplication.EndBlockCalls.ExpectCall(1)
	s.node.testApplication.EndBlockCalls.WaitForExpected()

	res := new(CurrentHeightResult)
	_, err := s.cli.Call("current_height", map[string]interface{}{}, res)
	c.Assert(err, IsNil)
	c.Check(res.Height, Equals, s.node.proxy.lastHeight)

}

func (s *RPCSuite) TestCanChangeValidators(c *C) {
	s.node.testApplication.EndBlockCalls.ExpectCall(2)
	s.node.testApplication.EndBlockCalls.WaitForExpected()

	// change should be in the future
	res := new(ChangeValidatorsResult)
	_, err := s.cli.Call("change_validators", map[string]interface{}{
		"scheduled_height": s.node.proxy.lastHeight - 1,
		"validators":       []*types.Validator{},
	}, res)
	c.Check(err, ErrorMatches, `Response error: Could not schedule for a block height back in time.*`)

	_, err = s.cli.Call("change_validators", map[string]interface{}{
		"scheduled_height": s.node.proxy.lastHeight + 5,
		"validators": []*types.Validator{
			&types.Validator{
				PubKey: s.genesisFile.Validators[0].PubKey.Bytes(),
				Power:  20,
			},
		},
	}, res)
	c.Check(err, IsNil)

}

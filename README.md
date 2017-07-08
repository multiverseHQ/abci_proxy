# ABCI Proxy

## Role and purpose

The role of ABCI proxy is to separate the
[ABCI interface](https://github.com/tendermint/abci/tree/master/client)
into two set of interface. One, internal for the ABCI 1-Click project
to manage the set of validator and observer on the fly via a
web-interface, and another set to pass-by request and response between
tendermint and the target client app.


## RPC Remote calls

There are two RPC calls implemented by the proxy (default listening port `46660` )

### Method `current_height`

* params: none
* results: 
  * Height : the current height
 
#### Example JSON request

```json
{
	"method": "current_height",
	"jsonrpc": "2.0",
	"params": [],
	"id": "dontcare"
}
```

#### Example JSON response

```json
{
	"jsonrpc": "2.0",
	"id": "dontcare",
	"result": {
		"height" : 1234
	},
	"error": ""
}
```

### Method `change_validators`

* params: 
  * `validators`: the list of validators to change
  * `scheduled_height` : the scheduled height (should be higher than current_height
*  results:  none

#### example JSON request

```json
{
	"method": "change_validators",
	"jsonrpc": "2.0",
	"params": {
		"scheduled_height": 1234,
		"validators":[
		{
			"pubKey": {
				"type" : "<TYPE>",
				"data" : "<HEXDATA>"
			},
			"power" : 10
		},
		{
			"pubKey": {
				"type" : "<TYPE>",
				"data" : "<HEXDATA>"
			},
			"power" : 10
		}
		]
	},
	"id": "dontcare"
}
```

#### Example JSON response

```json
{
	"jsonrpc": "2.0",
	"id": "dontcare",
	"result": {},
	"error": ""
}
```

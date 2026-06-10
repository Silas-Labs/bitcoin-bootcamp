package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type rpcRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      string `json:"id"`
	Method  string `json:"method"`
	Params  []any  `json:"params"`
}

type rpcResponse struct {
	Result json.RawMessage `json:"result"`
	Error  *struct {
		Message string `json:"message"`
	} `json:"error"`
}

var (
	rpcURL      = "http://127.0.0.1:18443/"
	rpcUser     = "bootcamp"
	rpcPassword = "bootcamp"
)

type BlockchainInfo struct {
	Chain      string  `json:"chain"`
	Blocks     int     `json:"blocks"`
	Difficulty float64 `json:"difficulty"`
}

type Transaction struct {
	Address          string   `json:"address,omitempty"`
	ParentDescs      []string `json:"parent_descs,omitempty"`
	Category         string   `json:"category"`
	Amount           float64  `json:"amount"`
	Label            string   `json:"label,omitempty"`
	Vout             int      `json:"vout,omitempty"`
	Abandoned        bool     `json:"abandoned,omitempty"`
	Confirmations    int      `json:"confirmations,omitempty"`
	Generated        bool     `json:"generated,omitempty"`
	BlockHash        string   `json:"blockhash,omitempty"`
	BlockHeight      int      `json:"blockheight,omitempty"`
	BlockIndex       int      `json:"blockindex,omitempty"`
	BlockTime        int64    `json:"blocktime,omitempty"`
	TxID             string   `json:"txid"`
	WTxID            string   `json:"wtxid,omitempty"`
	WalletConflicts  []string `json:"walletconflicts,omitempty"`
	MempoolConflicts []string `json:"mempoolconflicts,omitempty"`
	Time             int64    `json:"time,omitempty"`
	TimeReceived     int64    `json:"timereceived,omitempty"`
}

type Vin struct {
	TxID string `json:"txid,omitempty"`
	Vout int    `json:"vout,omitempty"`

	Coinbase string `json:"coinbase,omitempty"`
	Sequence uint32 `json:"sequence"`
}

type RawTransaction struct {
	TxID string `json:"txid"`
	Hash string `json:"hash"`
	Vin  []Vin  `json:"vin"`
	Vout []Vout `json:"vout"`
}
type Vout struct {
	Value        float64      `json:"value"`
	N            int          `json:"n"`
	ScriptPubKey ScriptPubKey `json:"scriptPubKey"`
}

type ScriptPubKey struct {
	Type    string `json:"type"`
	Address string `json:"address,omitempty"`
	Asm     string `json:"asm"`
	Hex     string `json:"hex"`
	Desc    string `json:"desc,omitempty"`
}

type Block struct {
	Hash          string   `json:"hash"`
	Confirmations int      `json:"confirmations"`
	Height        int      `json:"height"`
	Time          int64    `json:"time"`
	Tx            []string `json:"tx"`
}

func main() {
	showBlockchainInfo()
	showWalletBalance("alice")
	listTransactions("alice", 7)
	decodeTransaction("294d35f47ee1347768cd237158273f5c252990cdb958a05993e98f81d33a1aec")
	showBlock("2028654721cf104514a1b68e4328510493cbd9194d39444e207b9cd151ce777c")
}

func showBlockchainInfo() error {
	var info BlockchainInfo

	err := rpc("getblockchaininfo", nil, "", &info)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Chain: %s\n", info.Chain)
	fmt.Printf("Blocks: %d\n", info.Blocks)
	fmt.Printf("Blocks: %v\n", info.Difficulty)
	return nil
}

func showWalletBalance(wallet string) error {
	var balance float64

	if err := rpc("loadwallet", []any{wallet}, "", nil); err != nil {
		//  fmt.Printf("loadwalletError: %v\n", err)
	}

	err := rpc("getbalance", nil, wallet, &balance)
	if err != nil {
		fmt.Printf("getbalance failed: %v\n", err)
		return err
	}

	fmt.Printf("==== Wallet: %s ====\n", wallet)
	fmt.Printf("Balance: %v BTC\n", balance)
	return nil
}

func listTransactions(wallet string, count int) error {
	rpc("loadwallet", []any{wallet}, "", nil)
	var transactions []Transaction
	err := rpc("listtransactions", []any{"*", count}, "", &transactions)
	if err != nil {
		fmt.Println(err)
		return err
	}
	for i, v := range transactions {
		fmt.Printf("\n=== Transaction: %v ====\n", i+1)
		fmt.Printf("Direction: %v\n", v.Category)
		fmt.Printf("Amount: %v\n", v.Amount)
		fmt.Printf("TxID: %v\n", v.TxID)
	}
	return nil
}

func decodeTransaction(txid string) error {
	var tx RawTransaction

	if err := rpc("getrawtransaction", []any{txid, true}, "", &tx); err != nil {
		fmt.Println(err)
		return err
	}

	fmt.Println("===== DECODE =====")
	for _, vin := range tx.Vin {
		if vin.Coinbase != "" {
			fmt.Println("  COINBASE {mining reward}")
		} else {
			fmt.Printf("  From: %s...\n", vin.TxID[:20])
		}

	}

	for _, vout := range tx.Vout {
		fmt.Printf(" %.8f BTC -> %s\n", vout.Value, vout.ScriptPubKey.Address)
	}

	return nil
}

func rpc(method string, params []any, wallet string, out any) error {
	base := strings.TrimRight(rpcURL, "/")
	url := ""
	if wallet != "" {
		url = base + "/wallet/" + wallet
	} else {
		url = base
	}

	reqBody, err := json.Marshal(rpcRequest{
		JSONRPC: "1.0",
		ID:      "explorer",
		Method:  method,
		Params:  params,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return err
	}

	req.SetBasicAuth(rpcUser, rpcPassword)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("rpc status %d: %s", resp.StatusCode, string(body))
	}

	var parsed rpcResponse

	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return fmt.Errorf("decode rpc response: %w", err)
	}

	if parsed.Error != nil {
		return fmt.Errorf("rpc error: %s", parsed.Error.Message)
	}

	if len(parsed.Result) == 0 {
		return fmt.Errorf("rpc returned empty result")
	}

	if out != nil {
		if err := json.Unmarshal(parsed.Result, out); err != nil {
			return fmt.Errorf("unmarshal result: %w", err)
		}
	}
	return nil
}

func showBlock(blockhash string) error {
	if blockhash == "" {
		if err := rpc("getbestblockhash", nil, "", &blockhash); err != nil {
			return err
		}
	}

	var block Block

	if err := rpc("getblock", []any{blockhash, 1}, "", &block); err != nil {
		return err
	}

	fmt.Printf("Height: %d\n", block.Height)
	fmt.Printf("Hash:   %s\n", block.Hash)
	fmt.Printf("Time:   %s\n", time.Unix(block.Time, 0).Format(time.RFC3339))
	fmt.Printf("TXs:    %d\n", len(block.Tx))

	return nil
}

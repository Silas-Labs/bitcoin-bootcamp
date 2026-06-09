package main

import (
	"bytes"
    "encoding/json"
    "fmt"
    "net/http"
	"io"
)
type rpcRequest struct{
		JSONRPC string `json:"jsonrpc"`
		ID string `json:"id"`
		Method string `json:"method"`
		Params []any `json:"params"`
	}

	type rpcResponse struct{
		Result json.RawMessage `json:"result"`
		Error *struct{
			Message string `json:"message"`
		} `json:"error"`
	}

	var (
    rpcURL      = "http://127.0.0.1:18443/"
    rpcUser     = "bootcamp"
    rpcPassword = "bootcamp"
)
type BlockchainInfo struct {
    Chain  string `json:"chain"`
    Blocks int    `json:"blocks"`
	Difficulty float64 `json:"difficulty"`
}


func main() {
    showBlockchainInfo()
	fmt.Println(showWalletBalance("alice"))
}

func showBlockchainInfo()error{
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

func showWalletBalance(wallet string)error{
var balance float64

// optional: only if needed
_ = rpc("loadwallet", []any{wallet}, "", nil)

err := rpc("getbalance", nil, wallet, &balance)
if err != nil {
    return err
}

fmt.Printf("Balance: %f\n", balance)
return nil
}

func rpc(method string, params []any, wallet string, out any) error {
	url := rpcURL
	if wallet != "" {
		url += "/wallet/" + wallet
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

	req.SetBasicAuth(rpcUser,rpcPassword)
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

	if err := json.Unmarshal(parsed.Result, out); err != nil {
		return fmt.Errorf("unmarshal result: %w", err)
	}

	return nil
}
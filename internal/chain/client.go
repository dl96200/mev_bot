package chain

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	rpcURL string
	http   *http.Client
}

type rpcRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      int         `json:"id"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result"`
	Error   *rpcError       `json:"error"`
	ID      int             `json:"id"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func NewClient(rpcURL string) *Client {
	return &Client{
		rpcURL: rpcURL,
		http: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) Call(ctx context.Context, method string, params interface{}, out interface{}) error {
	payload := rpcRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      1,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal rpc payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.rpcURL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("create rpc request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("rpc request failed: %w", err)
	}
	defer resp.Body.Close()

	var decoded rpcResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return fmt.Errorf("decode rpc response: %w", err)
	}
	if decoded.Error != nil {
		return fmt.Errorf("rpc error %d: %s", decoded.Error.Code, decoded.Error.Message)
	}
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(decoded.Result, out); err != nil {
		return fmt.Errorf("decode rpc result: %w", err)
	}
	return nil
}

func (c *Client) URL() string {
	return c.rpcURL
}

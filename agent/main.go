package main

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/amshithnair/zcrypt/crypto"
	"github.com/amshithnair/zcrypt/utils"
)

const DEFAULT_SERVER = "http://localhost:8080"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	switch os.Args[1] {
	case "genkey":
		handleGenKey()
	case "log":
		handleLog()
	case "verify":
		handleVerify()
	case "chain-verify":
		handleChainVerify()
	case "chain-stats":
		handleChainStats()
	case "chain-export":
		handleChainExport()
	case "send-to-server":
		handleSendToServer()
	case "server-stats":
		handleServerStats()
	case "server-verify":
		handleServerVerify()
	case "register-agent":
		handleRegisterAgent()
	default:
		fmt.Println("Unknown command:", os.Args[1])
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Zcrypt - Cryptographic Log Chain CLI")
	fmt.Println("\nLocal Commands:")
	fmt.Println("  zcrypt genkey                          - Generate a keypair")
	fmt.Println("  zcrypt log \"message\"                   - Sign and store log entry locally")
	fmt.Println("  zcrypt verify \"message\" <signature>    - Verify a log signature")
	fmt.Println("  zcrypt chain-verify                    - Verify entire local log chain")
	fmt.Println("  zcrypt chain-stats                     - Show local chain statistics")
	fmt.Println("  zcrypt chain-export                    - Export local chain as JSON")
	fmt.Println("\nServer Commands:")
	fmt.Println("  zcrypt send-to-server \"message\"        - Send log to central server")
	fmt.Println("  zcrypt server-stats                    - Get server statistics")
	fmt.Println("  zcrypt server-verify                   - Verify server chain integrity")
	fmt.Println("  zcrypt register-agent <id> <name>      - Register this agent with server")
}

func handleGenKey() {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		fmt.Println("Error generating key:", err)
		return
	}

	os.MkdirAll(os.Getenv("HOME")+"/.zcrypt", 0700)
	
	err = os.WriteFile("zcrypt_private.key", priv, 0600)
	if err != nil {
		fmt.Println("Error saving private key:", err)
		return
	}

	err = os.WriteFile("zcrypt_public.key", pub, 0644)
	if err != nil {
		fmt.Println("Error saving public key:", err)
		return
	}

	fmt.Println("✓ Keypair generated successfully!")
	fmt.Println("Private key: zcrypt_private.key")
	fmt.Println("Public key: zcrypt_public.key")
	fmt.Printf("Public key hex: %s\n", hex.EncodeToString(pub))
}

func handleLog() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: zcrypt log \"message\"")
		return
	}

	message := os.Args[2]
	privKey, err := os.ReadFile("zcrypt_private.key")
	if err != nil {
		fmt.Println("Error: Private key not found. Run 'zcrypt genkey' first.")
		return
	}

	pubKey, err := os.ReadFile("zcrypt_public.key")
	if err != nil {
		fmt.Println("Error: Public key not found.")
		return
	}

	signature := ed25519.Sign(ed25519.PrivateKey(privKey), []byte(message))
	sigHex := hex.EncodeToString(signature)

	chainPath := crypto.GetChainPath()
	os.MkdirAll(os.Getenv("HOME")+"/.zcrypt", 0700)
	
	chain, err := crypto.NewLogChain(chainPath)
	if err != nil {
		fmt.Println("Error initializing chain:", err)
		return
	}

	hostname, _ := os.Hostname()
	entry, err := chain.AddLog(
		message,
		sigHex,
		hex.EncodeToString(pubKey),
		map[string]interface{}{
			"user":     os.Getenv("USER"),
			"hostname": hostname,
		},
	)
	if err != nil {
		fmt.Println("Error adding log:", err)
		return
	}

	fmt.Println("✓ Log entry added to local chain")
	fmt.Printf("  Message: %s\n", message)
	fmt.Printf("  Signature: %s\n", sigHex[:32]+"...")
	fmt.Printf("  Hash: %s\n", entry.CurrentHash[:32]+"...")
	fmt.Printf("  Prev Hash: %s\n", entry.PrevHash[:min(len(entry.PrevHash), 32)]+"...")
	fmt.Printf("  Chain length: %d\n", len(chain.Entries))
}

func handleVerify() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: zcrypt verify \"message\" <signature>")
		return
	}

	message := os.Args[2]
	sigHex := os.Args[3]

	pubKey, err := os.ReadFile("zcrypt_public.key")
	if err != nil {
		fmt.Println("Error: Public key not found.")
		return
	}

	signature, err := hex.DecodeString(sigHex)
	if err != nil {
		fmt.Println("Error: Invalid signature format")
		return
	}

	valid := ed25519.Verify(ed25519.PublicKey(pubKey), []byte(message), signature)
	
	if valid {
		fmt.Println("✓ Signature is VALID")
	} else {
		fmt.Println("✗ Signature is INVALID")
	}
}

func handleChainVerify() {
	chainPath := crypto.GetChainPath()
	chain, err := crypto.NewLogChain(chainPath)
	if err != nil {
		fmt.Println("Error loading chain:", err)
		return
	}

	valid, errors := chain.VerifyChain()
	
	if valid {
		fmt.Println("✓ Chain integrity verified - all hashes valid!")
		fmt.Printf("  Total entries: %d\n", len(chain.Entries))
	} else {
		fmt.Println("✗ Chain integrity COMPROMISED!")
		fmt.Println("  Errors found:")
		for _, e := range errors {
			fmt.Printf("    - %s\n", e)
		}
	}
}

func handleChainStats() {
	chainPath := crypto.GetChainPath()
	chain, err := crypto.NewLogChain(chainPath)
	if err != nil {
		fmt.Println("Error loading chain:", err)
		return
	}

	stats := chain.Stats()
	
	fmt.Println("Local Chain Statistics:")
	fmt.Printf("  Total entries: %d\n", stats["total_entries"])
	fmt.Printf("  Last hash: %s\n", stats["last_hash"].(string)[:min(len(stats["last_hash"].(string)), 64)])
	
	if stats["first_timestamp"] != nil {
		fmt.Printf("  First entry: %s\n", stats["first_timestamp"].(time.Time).Format("2006-01-02 15:04:05"))
		fmt.Printf("  Last entry: %s\n", stats["last_timestamp"].(time.Time).Format("2006-01-02 15:04:05"))
	}

	if len(chain.Entries) > 0 {
		fmt.Println("\nRecent entries (last 5):")
		start := max(0, len(chain.Entries)-5)
		for i := start; i < len(chain.Entries); i++ {
			entry := chain.Entries[i]
			fmt.Printf("  [%d] %s - %s\n", 
				i+1, 
				entry.Timestamp.Format("15:04:05"),
				entry.Message)
		}
	}
}

func handleChainExport() {
	chainPath := crypto.GetChainPath()
	chain, err := crypto.NewLogChain(chainPath)
	if err != nil {
		fmt.Println("Error loading chain:", err)
		return
	}

	json, err := chain.ExportJSON()
	if err != nil {
		fmt.Println("Error exporting chain:", err)
		return
	}

	exportPath := "zcrypt_chain_export.json"
	err = os.WriteFile(exportPath, []byte(json), 0644)
	if err != nil {
		fmt.Println("Error saving export:", err)
		return
	}

	fmt.Printf("✓ Chain exported to: %s\n", exportPath)
	fmt.Printf("  Total entries: %d\n", len(chain.Entries))
}

func handleSendToServer() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: zcrypt send-to-server \"message\"")
		return
	}

	message := os.Args[2]
	
	// Get server URL from env or use default
	serverURL := os.Getenv("ZCRYPT_SERVER")
	if serverURL == "" {
		serverURL = DEFAULT_SERVER
	}

	// Load keys
	privKey, err := os.ReadFile("zcrypt_private.key")
	if err != nil {
		fmt.Println("Error: Private key not found. Run 'zcrypt genkey' first.")
		return
	}

	pubKey, err := os.ReadFile("zcrypt_public.key")
	if err != nil {
		fmt.Println("Error: Public key not found.")
		return
	}

	// Sign message
	signature := ed25519.Sign(ed25519.PrivateKey(privKey), []byte(message))
	sigHex := hex.EncodeToString(signature)
	pubKeyHex := hex.EncodeToString(pubKey)

	// Get agent ID
	hostname, _ := os.Hostname()
	agentID := fmt.Sprintf("%s-%s", os.Getenv("USER"), hostname)

	// Create client
	client := utils.NewLogClient(serverURL)

	// Check server health
	healthy, err := client.HealthCheck()
	if err != nil || !healthy {
		fmt.Printf("Error: Cannot reach server at %s\n", serverURL)
		return
	}

	// Submit log
	submission := utils.LogSubmission{
		Message:   message,
		Signature: sigHex,
		PubKey:    pubKeyHex,
		AgentID:   agentID,
		Metadata: map[string]interface{}{
			"user":     os.Getenv("USER"),
			"hostname": hostname,
		},
	}

	resp, err := client.SubmitLog(submission)
	if err != nil {
		fmt.Println("Error submitting log:", err)
		return
	}

	fmt.Println("✓ Log submitted to server successfully!")
	fmt.Printf("  Server URL: %s\n", serverURL)
	fmt.Printf("  Chain length on server: %d\n", resp.ChainLength)
}

func handleServerStats() {
	serverURL := os.Getenv("ZCRYPT_SERVER")
	if serverURL == "" {
		serverURL = DEFAULT_SERVER
	}

	client := utils.NewLogClient(serverURL)
	stats, err := client.GetStats()
	if err != nil {
		fmt.Println("Error getting server stats:", err)
		return
	}

	fmt.Println("Server Statistics:")
	for key, value := range stats {
		fmt.Printf("  %s: %v\n", key, value)
	}
}

func handleServerVerify() {
	serverURL := os.Getenv("ZCRYPT_SERVER")
	if serverURL == "" {
		serverURL = DEFAULT_SERVER
	}

	client := utils.NewLogClient(serverURL)
	resp, err := client.VerifyChain()
	if err != nil {
		fmt.Println("Error verifying server chain:", err)
		return
	}

	if resp.Data != nil {
		valid, _ := resp.Data["valid"].(bool)
		if valid {
			fmt.Println("✓ Server chain integrity verified!")
			fmt.Printf("  Total entries: %.0f\n", resp.Data["total"].(float64))
		} else {
			fmt.Println("✗ Server chain integrity COMPROMISED!")
			if errors, ok := resp.Data["errors"].([]interface{}); ok {
				fmt.Println("  Errors found:")
				for _, e := range errors {
					fmt.Printf("    - %s\n", e)
				}
			}
		}
	}
}

func handleRegisterAgent() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: zcrypt register-agent <agent_id> <name>")
		return
	}

	agentID := os.Args[2]
	name := os.Args[3]

	serverURL := os.Getenv("ZCRYPT_SERVER")
	if serverURL == "" {
		serverURL = DEFAULT_SERVER
	}

	pubKey, err := os.ReadFile("zcrypt_public.key")
	if err != nil {
		fmt.Println("Error: Public key not found. Run 'zcrypt genkey' first.")
		return
	}

	client := utils.NewLogClient(serverURL)
	err = client.RegisterAgent(agentID, hex.EncodeToString(pubKey), name)
	if err != nil {
		fmt.Println("Error registering agent:", err)
		return
	}

	fmt.Println("✓ Agent registered successfully!")
	fmt.Printf("  Agent ID: %s\n", agentID)
	fmt.Printf("  Name: %s\n", name)
	fmt.Printf("  Server: %s\n", serverURL)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
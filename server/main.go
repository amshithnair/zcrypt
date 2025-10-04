// server/main.go
package main

import (
	"crypto/ed25519"
	"encoding/hex"
	"log"
	"time"
	"fmt"

	"github.com/amshithnair/zcrypt/crypto"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

type ServerConfig struct {
	Port       string
	ChainPath  string
	LogChain   *crypto.LogChain
	PubKeyRepo map[string]ed25519.PublicKey // agent_id -> public key
}

var config *ServerConfig

func main() {
	// Initialize server config
	config = &ServerConfig{
		Port:       ":8080",
		ChainPath:  "./server_logs.chain",
		PubKeyRepo: make(map[string]ed25519.PublicKey),
	}

	// Initialize server-side log chain
	chain, err := crypto.NewLogChain(config.ChainPath)
	if err != nil {
		log.Fatal("Failed to initialize chain:", err)
	}
	config.LogChain = chain

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "Zcrypt Log Server v1.0",
		ServerHeader: "Zcrypt",
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path} (${latency})\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// Routes
	setupRoutes(app)

	// Start server
	log.Printf("🚀 Zcrypt Server starting on http://localhost%s", config.Port)
	log.Fatal(app.Listen(config.Port))
}

func setupRoutes(app *fiber.App) {
	api := app.Group("/api/v1")

	// Health check
	api.Get("/health", healthCheck)

	// Log management
	logs := api.Group("/logs")
	logs.Post("/", submitLog)
	logs.Get("/", getLogs)
	logs.Get("/:id", getLogById)
	logs.Get("/range", getLogsByRange)

	// Verification
	verify := api.Group("/verify")
	verify.Post("/signature", verifySignature)
	verify.Post("/chain", verifyChain)

	// Agent management
	agents := api.Group("/agents")
	agents.Post("/register", registerAgent)
	agents.Get("/", listAgents)

	// Stats
	api.Get("/stats", getStats)
}

// Health check endpoint
func healthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "ok",
		"service": "zcrypt-server",
		"time":    time.Now().UTC(),
	})
}

// Submit a new log entry
func submitLog(c *fiber.Ctx) error {
	type LogRequest struct {
		Message   string                 `json:"message"`
		Signature string                 `json:"signature"`
		PubKey    string                 `json:"pubkey"`
		AgentID   string                 `json:"agent_id"`
		Metadata  map[string]interface{} `json:"metadata,omitempty"`
	}

	var req LogRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate required fields
	if req.Message == "" || req.Signature == "" || req.PubKey == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Missing required fields: message, signature, pubkey",
		})
	}

	// Verify signature
	pubKeyBytes, err := hex.DecodeString(req.PubKey)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid public key format",
		})
	}

	sigBytes, err := hex.DecodeString(req.Signature)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid signature format",
		})
	}

	valid := ed25519.Verify(ed25519.PublicKey(pubKeyBytes), []byte(req.Message), sigBytes)
	if !valid {
		return c.Status(401).JSON(fiber.Map{
			"error": "Invalid signature - verification failed",
		})
	}

	// Add metadata
	if req.Metadata == nil {
		req.Metadata = make(map[string]interface{})
	}
	req.Metadata["agent_id"] = req.AgentID
	req.Metadata["server_received"] = time.Now().UTC()

	// Add to chain
	entry, err := config.LogChain.AddLog(req.Message, req.Signature, req.PubKey, req.Metadata)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to add log to chain",
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"success":      true,
		"entry":        entry,
		"chain_length": len(config.LogChain.Entries),
	})
}

// Get all logs
func getLogs(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 100)
	offset := c.QueryInt("offset", 0)

	entries := config.LogChain.Entries
	total := len(entries)

	// Pagination
	if offset >= total {
		return c.JSON(fiber.Map{
			"entries": []crypto.LogEntry{},
			"total":   total,
			"limit":   limit,
			"offset":  offset,
		})
	}

	end := offset + limit
	if end > total {
		end = total
	}

	return c.JSON(fiber.Map{
		"entries": entries[offset:end],
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

// Get log by index
func getLogById(c *fiber.Ctx) error {
	id := c.Params("id")
	
	// Convert string to int manually
	index := 0
	if _, err := fmt.Sscanf(id, "%d", &index); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid log ID - must be a number",
		})
	}

	entry, err := config.LogChain.GetEntry(index)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Log entry not found",
		})
	}

	return c.JSON(fiber.Map{
		"entry": entry,
	})
}

// Get logs by time range
func getLogsByRange(c *fiber.Ctx) error {
	startStr := c.Query("start")
	endStr := c.Query("end")

	if startStr == "" || endStr == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Missing start or end time (RFC3339 format)",
		})
	}

	start, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid start time format (use RFC3339)",
		})
	}

	end, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid end time format (use RFC3339)",
		})
	}

	entries := config.LogChain.GetEntriesRange(start, end)

	return c.JSON(fiber.Map{
		"entries": entries,
		"count":   len(entries),
		"start":   start,
		"end":     end,
	})
}

// Verify a signature
func verifySignature(c *fiber.Ctx) error {
	type VerifyRequest struct {
		Message   string `json:"message"`
		Signature string `json:"signature"`
		PubKey    string `json:"pubkey"`
	}

	var req VerifyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	pubKeyBytes, err := hex.DecodeString(req.PubKey)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid public key format",
		})
	}

	sigBytes, err := hex.DecodeString(req.Signature)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid signature format",
		})
	}

	valid := ed25519.Verify(ed25519.PublicKey(pubKeyBytes), []byte(req.Message), sigBytes)

	return c.JSON(fiber.Map{
		"valid":   valid,
		"message": req.Message,
		"pubkey":  req.PubKey,
	})
}

// Verify chain integrity
func verifyChain(c *fiber.Ctx) error {
	valid, errors := config.LogChain.VerifyChain()

	return c.JSON(fiber.Map{
		"valid":  valid,
		"errors": errors,
		"total":  len(config.LogChain.Entries),
	})
}

// Register an agent
func registerAgent(c *fiber.Ctx) error {
	type RegisterRequest struct {
		AgentID string `json:"agent_id"`
		PubKey  string `json:"pubkey"`
		Name    string `json:"name,omitempty"`
	}

	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.AgentID == "" || req.PubKey == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Missing agent_id or pubkey",
		})
	}

	pubKeyBytes, err := hex.DecodeString(req.PubKey)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid public key format",
		})
	}

	config.PubKeyRepo[req.AgentID] = ed25519.PublicKey(pubKeyBytes)

	return c.Status(201).JSON(fiber.Map{
		"success":  true,
		"agent_id": req.AgentID,
		"message":  "Agent registered successfully",
	})
}

// List all registered agents
func listAgents(c *fiber.Ctx) error {
	agents := make([]fiber.Map, 0, len(config.PubKeyRepo))
	for agentID, pubKey := range config.PubKeyRepo {
		agents = append(agents, fiber.Map{
			"agent_id": agentID,
			"pubkey":   hex.EncodeToString(pubKey),
		})
	}

	return c.JSON(fiber.Map{
		"agents": agents,
		"count":  len(agents),
	})
}

// Get server statistics
func getStats(c *fiber.Ctx) error {
	stats := config.LogChain.Stats()
	stats["registered_agents"] = len(config.PubKeyRepo)

	return c.JSON(stats)
}
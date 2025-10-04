// crypto/chain.go
package crypto

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// LogEntry represents a single log in the chain
type LogEntry struct {
	Timestamp   time.Time              `json:"timestamp"`
	Message     string                 `json:"message"`
	Signature   string                 `json:"signature"`
	PubKey      string                 `json:"pubkey"`
	PrevHash    string                 `json:"prev_hash"`
	CurrentHash string                 `json:"current_hash"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// LogChain manages the immutable log ledger
type LogChain struct {
	Entries  []LogEntry `json:"entries"`
	FilePath string     `json:"-"`
	mu       sync.RWMutex
}

// NewLogChain initializes or loads existing chain
func NewLogChain(filePath string) (*LogChain, error) {
	lc := &LogChain{
		FilePath: filePath,
		Entries:  []LogEntry{},
	}

	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Try to load existing chain
	if _, err := os.Stat(filePath); err == nil {
		if err := lc.Load(); err != nil {
			return nil, fmt.Errorf("failed to load chain: %w", err)
		}
	}

	return lc, nil
}

// AddLog adds a new log entry to the chain
func (lc *LogChain) AddLog(message, signature, pubKey string, metadata map[string]interface{}) (*LogEntry, error) {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	// Get previous hash
	prevHash := "0" // Genesis block
	if len(lc.Entries) > 0 {
		prevHash = lc.Entries[len(lc.Entries)-1].CurrentHash
	}

	// Create new entry
	entry := LogEntry{
		Timestamp: time.Now().UTC(),
		Message:   message,
		Signature: signature,
		PubKey:    pubKey,
		PrevHash:  prevHash,
		Metadata:  metadata,
	}

	// Calculate current hash
	entry.CurrentHash = lc.calculateHash(entry)

	// Add to chain
	lc.Entries = append(lc.Entries, entry)

	// Persist to disk
	if err := lc.Save(); err != nil {
		return nil, fmt.Errorf("failed to save chain: %w", err)
	}

	return &entry, nil
}

// calculateHash computes SHA-256 hash of log entry
func (lc *LogChain) calculateHash(entry LogEntry) string {
	// Create deterministic string representation
	data := fmt.Sprintf("%s|%s|%s|%s|%s",
		entry.Timestamp.Format(time.RFC3339Nano),
		entry.Message,
		entry.Signature,
		entry.PubKey,
		entry.PrevHash,
	)

	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// VerifyChain checks integrity of entire chain
func (lc *LogChain) VerifyChain() (bool, []string) {
	lc.mu.RLock()
	defer lc.mu.RUnlock()

	var errors []string

	for i, entry := range lc.Entries {
		// Check hash
		expectedHash := lc.calculateHash(entry)
		if entry.CurrentHash != expectedHash {
			errors = append(errors, fmt.Sprintf("Entry %d: hash mismatch", i))
		}

		// Check chain linkage
		if i > 0 {
			if entry.PrevHash != lc.Entries[i-1].CurrentHash {
				errors = append(errors, fmt.Sprintf("Entry %d: broken chain link", i))
			}
		} else {
			if entry.PrevHash != "0" {
				errors = append(errors, "Entry 0: invalid genesis prev_hash")
			}
		}
	}

	return len(errors) == 0, errors
}

// GetLastHash returns the hash of the last entry
func (lc *LogChain) GetLastHash() string {
	lc.mu.RLock()
	defer lc.mu.RUnlock()

	if len(lc.Entries) == 0 {
		return "0"
	}
	return lc.Entries[len(lc.Entries)-1].CurrentHash
}

// GetEntry retrieves a specific log entry by index
func (lc *LogChain) GetEntry(index int) (*LogEntry, error) {
	lc.mu.RLock()
	defer lc.mu.RUnlock()

	if index < 0 || index >= len(lc.Entries) {
		return nil, fmt.Errorf("index out of range")
	}
	return &lc.Entries[index], nil
}

// GetEntriesRange retrieves logs within a time range
func (lc *LogChain) GetEntriesRange(start, end time.Time) []LogEntry {
	lc.mu.RLock()
	defer lc.mu.RUnlock()

	var result []LogEntry
	for _, entry := range lc.Entries {
		if (entry.Timestamp.Equal(start) || entry.Timestamp.After(start)) &&
			(entry.Timestamp.Equal(end) || entry.Timestamp.Before(end)) {
			result = append(result, entry)
		}
	}
	return result
}

// Save persists the chain to disk
func (lc *LogChain) Save() error {
	// Ensure directory exists before saving
	dir := filepath.Dir(lc.FilePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := json.MarshalIndent(lc, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	if err := os.WriteFile(lc.FilePath, data, 0600); err != nil {
		return fmt.Errorf("write error: %w", err)
	}

	return nil
}

// Load reads the chain from disk
func (lc *LogChain) Load() error {
	data, err := os.ReadFile(lc.FilePath)
	if err != nil {
		return fmt.Errorf("read error: %w", err)
	}

	if err := json.Unmarshal(data, lc); err != nil {
		return fmt.Errorf("unmarshal error: %w", err)
	}

	return nil
}

// ExportJSON exports chain to JSON string
func (lc *LogChain) ExportJSON() (string, error) {
	lc.mu.RLock()
	defer lc.mu.RUnlock()

	data, err := json.MarshalIndent(lc.Entries, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Stats returns chain statistics
func (lc *LogChain) Stats() map[string]interface{} {
	lc.mu.RLock()
	defer lc.mu.RUnlock()

	stats := map[string]interface{}{
		"total_entries": len(lc.Entries),
		"last_hash":     lc.GetLastHash(),
	}

	if len(lc.Entries) > 0 {
		stats["first_timestamp"] = lc.Entries[0].Timestamp
		stats["last_timestamp"] = lc.Entries[len(lc.Entries)-1].Timestamp
	}

	return stats
}

// GetChainPath returns the default chain file path
func GetChainPath() string {
	homeDir, _ := os.UserHomeDir()
	chainPath := filepath.Join(homeDir, ".zcrypt", "logs.chain")
	
	// Ensure .zcrypt directory exists
	zcryptDir := filepath.Join(homeDir, ".zcrypt")
	os.MkdirAll(zcryptDir, 0700)
	
	return chainPath
}
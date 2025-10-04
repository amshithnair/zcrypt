package crypto

import (
	"os"
	"testing"
	"time"
)

func TestNewLogChain(t *testing.T) {
	tempFile := os.TempDir() + "/test_chain.json"
	defer os.Remove(tempFile)

	chain, err := NewLogChain(tempFile)
	if err != nil {
		t.Fatalf("Failed to create chain: %v", err)
	}

	if len(chain.Entries) != 0 {
		t.Errorf("Expected empty chain, got %d entries", len(chain.Entries))
	}
}

func TestAddLog(t *testing.T) {
	tempFile := os.TempDir() + "/test_chain.json"
	defer os.Remove(tempFile)

	chain, _ := NewLogChain(tempFile)

	entry, err := chain.AddLog(
		"Test message",
		"test_signature",
		"test_pubkey",
		map[string]interface{}{"test": true},
	)

	if err != nil {
		t.Fatalf("Failed to add log: %v", err)
	}

	if entry.PrevHash != "0" {
		t.Errorf("Expected genesis prev_hash to be '0', got %s", entry.PrevHash)
	}

	if entry.CurrentHash == "" {
		t.Error("Expected current_hash to be set")
	}

	if len(chain.Entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(chain.Entries))
	}
}

func TestChainLinking(t *testing.T) {
	tempFile := os.TempDir() + "/test_chain.json"
	defer os.Remove(tempFile)

	chain, _ := NewLogChain(tempFile)

	entry1, _ := chain.AddLog("Log 1", "sig1", "key1", nil)
	entry2, _ := chain.AddLog("Log 2", "sig2", "key2", nil)
	entry3, _ := chain.AddLog("Log 3", "sig3", "key3", nil)

	if entry2.PrevHash != entry1.CurrentHash {
		t.Error("Entry 2 not linked to Entry 1")
	}

	if entry3.PrevHash != entry2.CurrentHash {
		t.Error("Entry 3 not linked to Entry 2")
	}
}

func TestVerifyChain(t *testing.T) {
	tempFile := os.TempDir() + "/test_chain.json"
	defer os.Remove(tempFile)

	chain, _ := NewLogChain(tempFile)

	chain.AddLog("Log 1", "sig1", "key1", nil)
	chain.AddLog("Log 2", "sig2", "key2", nil)

	valid, errors := chain.VerifyChain()
	if !valid {
		t.Errorf("Chain should be valid. Errors: %v", errors)
	}
}

func TestTamperedChain(t *testing.T) {
	tempFile := os.TempDir() + "/test_chain.json"
	defer os.Remove(tempFile)

	chain, _ := NewLogChain(tempFile)

	chain.AddLog("Log 1", "sig1", "key1", nil)
	chain.AddLog("Log 2", "sig2", "key2", nil)

	// Tamper with chain
	chain.Entries[0].Message = "TAMPERED MESSAGE"

	valid, errors := chain.VerifyChain()
	if valid {
		t.Error("Chain should be invalid after tampering")
	}

	if len(errors) == 0 {
		t.Error("Expected verification errors")
	}
}

func TestPersistence(t *testing.T) {
	tempFile := os.TempDir() + "/test_chain.json"
	defer os.Remove(tempFile)

	chain1, _ := NewLogChain(tempFile)
	chain1.AddLog("Log 1", "sig1", "key1", nil)
	chain1.AddLog("Log 2", "sig2", "key2", nil)

	chain2, err := NewLogChain(tempFile)
	if err != nil {
		t.Fatalf("Failed to load chain: %v", err)
	}

	if len(chain2.Entries) != 2 {
		t.Errorf("Expected 2 entries after reload, got %d", len(chain2.Entries))
	}

	if chain2.Entries[0].Message != "Log 1" {
		t.Error("Chain data not persisted correctly")
	}
}

func TestGetEntriesRange(t *testing.T) {
	tempFile := os.TempDir() + "/test_chain.json"
	defer os.Remove(tempFile)

	chain, _ := NewLogChain(tempFile)

	now := time.Now()
	chain.AddLog("Log 1", "sig1", "key1", nil)
	time.Sleep(10 * time.Millisecond)
	chain.AddLog("Log 2", "sig2", "key2", nil)
	time.Sleep(10 * time.Millisecond)
	chain.AddLog("Log 3", "sig3", "key3", nil)

	entries := chain.GetEntriesRange(now.Add(-1*time.Hour), now.Add(1*time.Hour))
	if len(entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(entries))
	}
}
package serialization

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gofast/internal/llm"
)

func TestHistoryManagement(t *testing.T) {
	t.Run("save and load history", func(t *testing.T) {
		// Create test memory with some messages
		memory := llm.NewSimpleMemory()
		messages := []llm.Message{
			{
				Type:    llm.MessageTypeUser,
				Content: "hello",
				Name:    "user1",
			},
			{
				Type:    llm.MessageTypeAssistant,
				Content: "hi there",
				Name:    "assistant1",
				Parts: []llm.MessagePart{
					{Type: "text", Content: "additional"},
				},
			},
		}

		for _, msg := range messages {
			require.NoError(t, memory.Add(msg, false))
		}

		// Create test metadata
		metadata := &HistoryMetadata{
			LLMName:  "test-llm",
			Provider: "test-provider",
			Model:    "test-model",
			Custom: map[string]interface{}{
				"test_key": "test_value",
			},
		}

		// Create temporary directory
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "history.yaml")

		// Save history
		err := SaveHistory(memory, path, metadata)
		require.NoError(t, err)

		// Create new memory for loading
		newMemory := llm.NewSimpleMemory()

		// Load history
		loadedMetadata, err := LoadHistory(newMemory, path, false)
		require.NoError(t, err)

		// Verify metadata
		assert.Equal(t, metadata.LLMName, loadedMetadata.LLMName)
		assert.Equal(t, metadata.Provider, loadedMetadata.Provider)
		assert.Equal(t, metadata.Model, loadedMetadata.Model)
		assert.Equal(t, metadata.Custom["test_key"], loadedMetadata.Custom["test_key"])

		// Verify messages
		loadedMessages, err := newMemory.Get(true)
		require.NoError(t, err)
		require.Len(t, loadedMessages, len(messages))

		for i, msg := range messages {
			assert.Equal(t, msg.Type, loadedMessages[i].Type)
			assert.Equal(t, msg.Content, loadedMessages[i].Content)
			assert.Equal(t, msg.Name, loadedMessages[i].Name)
			if len(msg.Parts) > 0 {
				assert.Equal(t, msg.Parts[0].Content, loadedMessages[i].Parts[0].Content)
			}
		}
	})

	t.Run("save to directory with timestamp", func(t *testing.T) {
		// Create test memory
		memory := llm.NewSimpleMemory()
		require.NoError(t, memory.Add(llm.Message{
			Type:    llm.MessageTypeUser,
			Content: "test message",
		}, false))

		// Create temporary directory
		tmpDir := t.TempDir()

		// Save history with timestamp
		path, err := SaveHistoryToDir(memory, tmpDir, "test", nil)
		require.NoError(t, err)

		// Verify file exists and has correct format
		_, err = os.Stat(path)
		require.NoError(t, err)

		filename := filepath.Base(path)
		assert.Contains(t, filename, "test.yaml")
		assert.Regexp(t, `^\d{8}_\d{6}_test\.yaml$`, filename)
	})

	t.Run("list history files", func(t *testing.T) {
		// Create temporary directory
		tmpDir := t.TempDir()

		// Create some test files
		files := []string{
			"20240101_120000_test1.yaml",
			"20240101_120001_test2.yaml",
			"not_a_history.txt",
		}

		for _, file := range files {
			path := filepath.Join(tmpDir, file)
			require.NoError(t, os.WriteFile(path, []byte("test"), 0644))
		}

		// List history files
		matches, err := ListHistoryFiles(tmpDir)
		require.NoError(t, err)

		// Should only find .yaml files
		assert.Len(t, matches, 2)
		for _, match := range matches {
			assert.Contains(t, match, ".yaml")
		}
	})

	t.Run("get history metadata", func(t *testing.T) {
		// Create test memory and metadata
		memory := llm.NewSimpleMemory()
		require.NoError(t, memory.Add(llm.Message{
			Type:    llm.MessageTypeUser,
			Content: "test",
		}, false))

		metadata := &HistoryMetadata{
			LLMName:  "test-llm",
			Provider: "test-provider",
			Model:    "test-model",
			Custom: map[string]interface{}{
				"test_key": "test_value",
			},
		}

		// Save history
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "history.yaml")
		require.NoError(t, SaveHistory(memory, path, metadata))

		// Get metadata without loading messages
		loadedMetadata, err := GetHistoryMetadata(path)
		require.NoError(t, err)

		// Verify metadata
		assert.Equal(t, metadata.LLMName, loadedMetadata.LLMName)
		assert.Equal(t, metadata.Provider, loadedMetadata.Provider)
		assert.Equal(t, metadata.Model, loadedMetadata.Model)
		assert.Equal(t, metadata.Custom["test_key"], loadedMetadata.Custom["test_key"])
	})

	t.Run("clear existing history", func(t *testing.T) {
		// Create test memory with existing messages
		memory := llm.NewSimpleMemory()
		require.NoError(t, memory.Add(llm.Message{
			Type:    llm.MessageTypeUser,
			Content: "existing message",
		}, false))

		// Save new history
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "history.yaml")
		require.NoError(t, SaveHistory(memory, path, nil))

		// Add different message to memory
		require.NoError(t, memory.Add(llm.Message{
			Type:    llm.MessageTypeUser,
			Content: "different message",
		}, false))

		// Load history with clearExisting=true
		_, err := LoadHistory(memory, path, true)
		require.NoError(t, err)

		// Verify only loaded messages exist
		messages, err := memory.Get(true)
		require.NoError(t, err)
		require.Len(t, messages, 1)
		assert.Equal(t, "existing message", messages[0].Content)
	})

	t.Run("error handling", func(t *testing.T) {
		memory := llm.NewSimpleMemory()
		tmpDir := t.TempDir()

		t.Run("nonexistent file", func(t *testing.T) {
			path := filepath.Join(tmpDir, "nonexistent.yaml")
			_, err := LoadHistory(memory, path, false)
			assert.Error(t, err)
		})

		t.Run("invalid directory", func(t *testing.T) {
			path := filepath.Join(os.DevNull, "test.yaml")
			err := SaveHistory(memory, path, nil)
			assert.Error(t, err)
		})

		t.Run("invalid yaml", func(t *testing.T) {
			path := filepath.Join(tmpDir, "invalid.yaml")
			err := os.WriteFile(path, []byte("invalid: ][yaml"), 0644)
			require.NoError(t, err)

			_, err = LoadHistory(memory, path, false)
			assert.Error(t, err)
		})
	})
}

func TestHistoryMetadata(t *testing.T) {
	t.Run("metadata serialization", func(t *testing.T) {
		now := time.Now()
		metadata := &HistoryMetadata{
			SavedAt:  now,
			LLMName:  "test-llm",
			Provider: "test-provider",
			Model:    "test-model",
			Custom: map[string]interface{}{
				"test_key": "test_value",
				"nested": map[string]interface{}{
					"key": "value",
				},
			},
		}

		// Create test memory
		memory := llm.NewSimpleMemory()
		require.NoError(t, memory.Add(llm.Message{
			Type:    llm.MessageTypeUser,
			Content: "test",
		}, false))

		// Save with metadata
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "history.yaml")
		require.NoError(t, SaveHistory(memory, path, metadata))

		// Load metadata
		loaded, err := GetHistoryMetadata(path)
		require.NoError(t, err)

		// Verify all fields
		assert.Equal(t, metadata.LLMName, loaded.LLMName)
		assert.Equal(t, metadata.Provider, loaded.Provider)
		assert.Equal(t, metadata.Model, loaded.Model)
		assert.Equal(t, metadata.Custom["test_key"], loaded.Custom["test_key"])
		assert.Equal(t,
			metadata.Custom["nested"].(map[string]interface{})["key"],
			loaded.Custom["nested"].(map[string]interface{})["key"],
		)
	})
}

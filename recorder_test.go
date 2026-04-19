package copilot

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/nuln/agent-core"
	"github.com/stretchr/testify/assert"
)

type mockKVStore struct {
	data map[string][]byte
}

func (m *mockKVStore) Get(key []byte) ([]byte, error)   { return m.data[string(key)], nil }
func (m *mockKVStore) Put(key, value []byte) error      { m.data[string(key)] = value; return nil }
func (m *mockKVStore) Delete(key []byte) error          { delete(m.data, string(key)); return nil }
func (m *mockKVStore) List() (map[string][]byte, error) { return m.data, nil }

type mockKVStoreProvider struct {
	store *mockKVStore
}

func (m *mockKVStoreProvider) GetStore(name string) (agent.KVStore, error) {
	return m.store, nil
}

func TestSessionRecorder_Table(t *testing.T) {
	tests := []struct {
		name     string
		events   []agent.Event
		validate func(t *testing.T, log copilotInteractionLog)
	}{
		{
			name: "full lifecycle",
			events: []agent.Event{
				{Type: agent.EventThinking, Content: "Thinking via Copilot..."},
				{Type: agent.EventToolUse, ToolName: "CopilotTool", ToolInput: "check health"},
				{Type: agent.EventText, Content: "Service is up."},
			},
			validate: func(t *testing.T, log copilotInteractionLog) {
				assert.Equal(t, "Thinking via Copilot...", log.Thinking)
				assert.Equal(t, "Service is up.", log.Response)
				assert.Len(t, log.ToolCalls, 1)
				assert.Equal(t, "CopilotTool", log.ToolCalls[0].Name)
			},
		},
		{
			name: "oversize content",
			events: []agent.Event{
				{Type: agent.EventText, Content: strings.Repeat("D", 10500)},
			},
			validate: func(t *testing.T, log copilotInteractionLog) {
				assert.Contains(t, log.Response, "[truncated")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &mockKVStore{data: make(map[string][]byte)}
			provider := &mockKVStoreProvider{store: store}

			sess := &copilotSession{
				model: "claude-3.5-sonnet",
				mode:  "autopilot",
			}
			sess.storage = provider
			sess.SetTraceID("t-copilot")

			for _, ev := range tt.events {
				sess.recordEvent(ev)
			}

			sess.finalizeInteractionLog()

			data, ok := store.data["t-copilot"]
			assert.True(t, ok)

			var log copilotInteractionLog
			_ = json.Unmarshal(data, &log)
			tt.validate(t, log)
		})
	}
}

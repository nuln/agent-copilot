package copilot

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/nuln/agent-core"
)

// copilotInteractionLog is Copilot's own data structure for recording interactions.
type copilotInteractionLog struct {
	TraceID   string               `json:"trace_id"`
	SessionID string               `json:"session_id"`
	Model     string               `json:"model"`
	Mode      string               `json:"mode"`
	WorkDir   string               `json:"work_dir"`
	Thinking  string               `json:"thinking,omitempty"`
	Response  string               `json:"response,omitempty"`
	ToolCalls []copilotToolCallLog `json:"tool_calls,omitempty"`
	StartTime int64                `json:"start_time"`
	EndTime   int64                `json:"end_time"`
}

// copilotToolCallLog records a single tool invocation.
type copilotToolCallLog struct {
	Name      string `json:"name"`
	Input     string `json:"input,omitempty"`
	Timestamp int64  `json:"ts"`
}

// SetStorage implements agent.StorageAware for the Copilot LLM plugin.
func (a *LLM) SetStorage(store agent.KVStoreProvider) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.storage = store
}

// SetTraceID implements agent.SessionRecorder for copilotSession.
func (cs *copilotSession) SetTraceID(traceID string) {
	cs.traceID = traceID
	cs.recordStartTime = time.Now().UnixMilli()
}

// recordEvent captures important events for the interaction log.
func (cs *copilotSession) recordEvent(ev agent.Event) {
	cs.logMu.Lock()
	defer cs.logMu.Unlock()

	if cs.traceID == "" {
		return
	}
	switch ev.Type {
	case agent.EventThinking:
		cs.recordedThinking.WriteString(ev.Content)
	case agent.EventText:
		cs.recordedResponse.WriteString(ev.Content)
	case agent.EventToolUse:
		cs.recordedTools = append(cs.recordedTools, copilotToolCallLog{
			Name:      ev.ToolName,
			Input:     agent.TruncateStr(ev.ToolInput, 500),
			Timestamp: time.Now().UnixMilli(),
		})
	}
}

// finalizeInteractionLog persists the complete interaction record to storage.
func (cs *copilotSession) finalizeInteractionLog() {
	cs.logMu.Lock()
	if cs.traceID == "" || cs.storage == nil {
		cs.logMu.Unlock()
		return
	}
	log := copilotInteractionLog{
		TraceID:   cs.traceID,
		SessionID: cs.currentSessionID(),
		Model:     cs.model,
		Mode:      cs.mode,
		WorkDir:   cs.workDir,
		Thinking:  agent.TruncateStr(cs.recordedThinking.String(), 10000),
		Response:  agent.TruncateStr(cs.recordedResponse.String(), 10000),
		ToolCalls: cs.recordedTools,
		StartTime: cs.recordStartTime,
		EndTime:   time.Now().UnixMilli(),
	}
	cs.logMu.Unlock()

	store, err := cs.storage.GetStore("interactions")
	if err != nil {
		slog.Warn("copilot: failed to get interaction store", "error", err)
		return
	}

	data, err := json.Marshal(log)
	if err != nil {
		slog.Warn("copilot: failed to marshal interaction log", "error", err)
		return
	}
	if err := store.Put([]byte(cs.traceID), data); err != nil {
		slog.Warn("copilot: failed to persist interaction log", "error", err, "trace_id", cs.traceID)
	}
}

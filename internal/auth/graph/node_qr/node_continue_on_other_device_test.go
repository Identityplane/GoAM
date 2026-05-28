package node_qr

import (
	"testing"

	"github.com/Identityplane/GoAM/pkg/model"
)

func TestRunContinueOnOtherDeviceNode(t *testing.T) {
	t.Parallel()

	const wantMsg = "Success — you can continue signing in on your computer. Your browser will finish signing in. Tap continue when you are ready to close this screen."

	tests := []struct {
		name       string
		input      map[string]string
		preSetFlag bool
		wantFlag   string
		wantCond   string
		wantPrompt string
	}{
		{
			name:       "first visit updates context then returns success message",
			input:      nil,
			preSetFlag: false,
			wantFlag:   "true",
			wantCond:   "",
			wantPrompt: wantMsg,
		},
		{
			name:       "ack completes without changing flag",
			input:      map[string]string{"ack": "true"},
			preSetFlag: true,
			wantFlag:   "true",
			wantCond:   "done",
			wantPrompt: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			state := &model.AuthenticationSession{
				Context: make(map[string]string),
			}
			if tt.preSetFlag {
				state.Context[CONTEXT_KEY_CONTINUE_ON_PRIMARY_DEVICE] = "true"
			}
			res, err := RunContinueOnOtherDeviceNode(state, &model.GraphNode{}, tt.input, nil)
			if err != nil {
				t.Fatalf("RunContinueOnOtherDeviceNode: %v", err)
			}
			if res.Condition != tt.wantCond {
				t.Fatalf("condition: got %q want %q", res.Condition, tt.wantCond)
			}
			got := ""
			if res.Prompts != nil {
				got = res.Prompts[continueOnOtherDevicePromptMessage]
			}
			if got != tt.wantPrompt {
				t.Fatalf("prompt message: got %q want %q", got, tt.wantPrompt)
			}
			if state.Context[CONTEXT_KEY_CONTINUE_ON_PRIMARY_DEVICE] != tt.wantFlag {
				t.Fatalf("flag: got %q want %q", state.Context[CONTEXT_KEY_CONTINUE_ON_PRIMARY_DEVICE], tt.wantFlag)
			}
		})
	}
}

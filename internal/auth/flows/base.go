package flows

type AuthLevel string

const (
	AuthLevelUnauthenticated AuthLevel = "0"
	AuthLevel1FA             AuthLevel = "1"
	AuthLevel2FA             AuthLevel = "2"
)

type FlowStep struct {
	Name       string            `json:"name"`       // Step name (e.g., "username", "password")
	Parameters map[string]string `json:"parameters"` // Input key/values for that step
}

type FlowState struct {
	RunID      string             `json:"run_id"`
	Steps      []FlowStep         `json:"steps"`            // History of steps taken
	Error      *string            `json:"error,omitempty"`  // Terminal error (if any)
	Result     *FlowResult        `json:"result,omitempty"` // Terminal success result (if any)
	UserInputs *map[string]string `json:"userInputs"`
	Variables  *map[string]string `json:"variables"`
}

type FlowResult struct {
	UserID        string    `json:"user_id"`
	Username      string    `json:"username"`
	Authenticated bool      `json:"authenticated"`
	AuthLevel     AuthLevel `json:"auth_level"`
}

type Flow interface {
	Run(state *FlowState) // updates the state in-place
}

func LastStep(state *FlowState) *FlowStep {
	if len(state.Steps) == 0 {
		return nil
	}
	return &state.Steps[len(state.Steps)-1]
}

func GetStep(state *FlowState, name string) *FlowStep {
	for _, step := range state.Steps {
		if step.Name == name {
			return &step
		}
	}
	return nil
}

func GetParam(state *FlowState, stepName, param string) (string, bool) {
	step := GetStep(state, stepName)
	if step == nil {
		return "", false
	}
	val, ok := step.Parameters[param]
	return val, ok
}

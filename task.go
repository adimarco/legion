package hive

import "fmt"

// Task represents a research task with assignments
type Task struct {
	question    string
	assignments map[string]string
}

// NewTask creates a new task with the given question
func NewTask(question string) *Task {
	return &Task{
		question:    question,
		assignments: make(map[string]string),
	}
}

// AssignTo adds an assignment for a team member
func (t *Task) AssignTo(member, question string) *Task {
	t.assignments[member] = question
	return t
}

// Run executes the task with the given team
func (t *Task) Run(team *Team) (map[string]string, error) {
	responses := make(map[string]string)
	for member, question := range t.assignments {
		response, err := team.Send(member, question)
		if err != nil {
			return nil, fmt.Errorf("failed to get response from %s: %w", member, err)
		}
		responses[member] = response
	}
	return responses, nil
}

// SynthesisRequest represents a request to synthesize findings
type SynthesisRequest struct {
	responses map[string]string
	prompt    string
}

// NewSynthesisRequest creates a new synthesis request
func NewSynthesisRequest() *SynthesisRequest {
	return &SynthesisRequest{}
}

// WithResponses adds responses to the synthesis request
func (s *SynthesisRequest) WithResponses(responses map[string]string) *SynthesisRequest {
	s.responses = responses
	return s
}

// WithPrompt sets the synthesis prompt
func (s *SynthesisRequest) WithPrompt(prompt string) *SynthesisRequest {
	s.prompt = prompt
	return s
}

// SendTo sends the synthesis request to the specified team member
func (s *SynthesisRequest) SendTo(team *Team, member string) (string, error) {
	summary := "Based on these findings:\n\n"
	for member, response := range s.responses {
		summary += fmt.Sprintf("%s found: %s\n\n", member, response)
	}
	summary += s.prompt
	return team.Send(member, summary)
}

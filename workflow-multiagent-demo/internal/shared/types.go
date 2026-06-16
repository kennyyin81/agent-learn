package shared

type CodeFile struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type ExecutionResult struct {
	Command    string `json:"command"`
	Dir        string `json:"dir"`
	Success    bool   `json:"success"`
	ExitCode   int    `json:"exit_code"`
	Stdout     string `json:"stdout"`
	Stderr     string `json:"stderr"`
	DurationMS int64  `json:"duration_ms"`
	TimedOut   bool   `json:"timed_out"`
}

package types

type TestCase struct {
	Stdin          string `json:"stdin"`
	ExpectedStdout string `json:"expected_stdout"`
}

type RunRequest struct {
	Language string     `json:"language"`
	Source   string     `json:"source"`
	Tests    []TestCase `json:"tests"`

	ResourceOverrides *ResourceOverrides `json:"resource_overrides,omitempty"`
	BuildFlags        []string           `json:"build_flags,omitempty"`
	RunFlags          []string           `json:"run_flags,omitempty"`
}

type ResourceOverrides struct {
	TimeLimitSeconds string `json:"time_limit_seconds,omitempty"`
	AddressSpaceMB   string `json:"address_space_mb,omitempty"`
	FileSizeMB       string `json:"file_size_mb,omitempty"`
	OpenFiles        string `json:"open_files,omitempty"`
	Processes        string `json:"processes,omitempty"`
}

type BuildResult struct {
	Status          string `json:"status"`
	Stdout          string `json:"stdout"`
	Stderr          string `json:"stderr"`
	StdoutTruncated bool   `json:"stdout_truncated"`
	StderrTruncated bool   `json:"stderr_truncated"`
	DurationMS      int64  `json:"duration_ms"`
}

type TestResult struct {
	Status          string `json:"status"`
	Stdout          string `json:"stdout"`
	Stderr          string `json:"stderr"`
	StdoutTruncated bool   `json:"stdout_truncated"`
	StderrTruncated bool   `json:"stderr_truncated"`
	DurationMS      int64  `json:"duration_ms"`
}

type RunResponse struct {
	Status string       `json:"status"`
	Build  *BuildResult `json:"build,omitempty"`
	Tests  []TestResult `json:"tests"`
}

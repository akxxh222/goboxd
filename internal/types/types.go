package types

type TestCase struct {
	Stdin          string `json:"stdin"`
	ExpectedStdout string `json:"expected_stdout"`
}

type RunRequest struct {
	Language string     `json:"language"`
	Source   string     `json:"source"`
	Tests    []TestCase `json:"tests"`
}

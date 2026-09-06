package main

import (
	"encoding/json"
	"io"
	"os/exec"
)

func runDoctor(out io.Writer) error {
	type tool struct {
		Name      string `json:"name"`
		Available bool   `json:"available"`
		Path      string `json:"path,omitempty"`
	}
	result := []tool{}
	for _, name := range []string{"git", "go", "node", "npm", "python", "rustc", "cargo", "java", "javac", "dotnet", "clang", "gcc"} {
		p, e := exec.LookPath(name)
		result = append(result, tool{name, e == nil, p})
	}
	return json.NewEncoder(out).Encode(struct {
		Tools []tool `json:"tools"`
		Note  string `json:"note"`
	}{result, "Availability only. No project commands, compilers, or indexers were executed."})
}

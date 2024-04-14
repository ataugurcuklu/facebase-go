package utils

import (
    "fmt"
    "os/exec"
    "path/filepath"
)

func CallPythonCLI(command string, args ...string) (string, error) {
    parentDir, err := filepath.Abs("..")
    if err != nil {
        return "", err
    }
    cmd := exec.Command(filepath.Join(parentDir, "env/Scripts/python"), append([]string{filepath.Join(parentDir, "cli.py"), command}, args...)...)
    out, err := cmd.Output()
    if err != nil {
        fmt.Println("Error:", err)
    }
    fmt.Println("Command:", cmd.String())
    return string(out), err
}
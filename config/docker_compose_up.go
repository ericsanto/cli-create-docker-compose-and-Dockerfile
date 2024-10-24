package config

import (
	"fmt"
	"os/exec"
)

func UpDockerCompose() {

	cmd := exec.Command("docker", "compose", "up", "-d")
	cmd.Dir = "."

	output, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Println("Erro ao executar comando", err)
	}

	fmt.Println(string(output))
}

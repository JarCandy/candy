package cli

import (
	"fmt"
	"os"
	"os/exec"
)

func Build() error {
	cmd := exec.Command("go", "build", "-o", "candy", "./cmd")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	fmt.Println("Built ./candy")
	return nil
}

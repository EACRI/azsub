package azsub

import "fmt"

func RunLocal(a *Azsub) error {

	if a.Task.IsCommand() {
		fmt.Printf("Running command task: %s\n", a.Task.Task)
	} else if a.Task.IsScript() {
		fmt.Printf("Running script task: %s\n", a.Task.Task)
	}

	return nil

}

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/jessevdk/go-flags"
)

const (
	agentK8sURL = "https://raw.githubusercontent.com/weaveworks/config/master/agent.yaml"
)

func main() {
	// Parse arguments with go-flags so we can forward unknown arguments to kubectl
	var opts struct {
		Token string `long:"token" description:"Weave Cloud token" required:"true"`
	}
	parser := flags.NewParser(&opts, flags.IgnoreUnknown)
	otherArgs, err := parser.Parse()
	if err != nil {
		die("%s\n", err)
	}

	// Ask the user to confirm the cluster
	cluster, err := getClusterInfo(otherArgs)
	if err != nil {
		die("There was an error fetching the current cluster info: %s\n", err)
	}

	fmt.Printf("\nThis will install Weave Cloud on the following cluster:\n")
	fmt.Printf("    Name: %s\n    Server: %s\n\n", cluster.Name, cluster.ServerAddress)
	fmt.Printf("Please run 'kubectl config use-context' or pass '--kubeconf' if you would like to change this.\n\n")

	confirmed, err := askForConfirmation("Would you like to continue?")
	if err != nil {
		die("There was an error: %s\n", err)
	}
	if !confirmed {
		fmt.Println("Cancelled.")
		return
	}

	fmt.Println("Storing the instance token in the weave-cloud secret...")
	_, err = executeKubectlCommand(
		append([]string{
			"create",
			"secret",
			"generic",
			"weave-cloud",
			fmt.Sprintf("--from-literal=token=%s", opts.Token),
		}, otherArgs...),
	)
	if err != nil {
		die("There was an error creating the secret: %s\n", err)
	}

	fmt.Println("Applying the agent...")
	_, err = executeKubectlCommand(append([]string{"apply", "-f", agentK8sURL}, otherArgs...))
	if err != nil {
		die("There was an error applying the agent: %s\n", err)
	}
}

func die(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg, args...)
	os.Exit(1)
}

func askForConfirmation(s string) (bool, error) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("%s [y/n]: ", s)
		response, err := reader.ReadString('\n')
		if err != nil {
			return false, err
		}

		response = strings.ToLower(strings.TrimSpace(response))
		if response == "y" || response == "yes" {
			return true, nil
		} else if response == "n" || response == "no" {
			return false, nil
		}
	}
}
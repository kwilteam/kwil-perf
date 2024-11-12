package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// This file can be used to generate the testnet configuration for all the nodes

var (
	vals      string
	nVals     string
	ipFile    string
	dir       string
	kwilAdmin string
)

func main() {
	flag.StringVar(&ipFile, "addresses", "ips.txt", "file containing the list of the IP Addresses of the nodes")

	flag.StringVar(&vals, "vals", "4", "number of validators")
	flag.StringVar(&nVals, "nvals", "0", "number of non-validators")
	flag.StringVar(&dir, "dir", ".testnet", "directory to store the configuration files")
	flag.StringVar(&kwilAdmin, "kwil-admin", "kwil-admin", "kwil-admin binary path")
	flag.Parse()

	// construct the configuration
	var args []string

	dirPath, err := expandPath(dir)
	if err != nil {
		fmt.Printf("Error expanding the directory path: %v", err)
		return
	}

	adminPath, err := expandPath(kwilAdmin)
	if err != nil {
		fmt.Printf("Error expanding the kwil-admin path: %v", err)
		return
	}

	ipFilePath, err := expandPath(ipFile)
	if err != nil {
		fmt.Printf("Error expanding the IP Address file path: %v", err)
		return
	}

	// construct the command arguments
	args = append(args, "setup", "testnet", "-v", vals, "-n", nVals, "-o", dirPath)

	fileR, err := os.Open(ipFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("IP Address file doesn't exist at location %s, please provide correct location: error %v", ipFilePath, err)
			return
		}
		fmt.Printf("Error reading the IP Address file: %v", err)
		return
	}
	defer fileR.Close()

	scanner := bufio.NewScanner(fileR)
	for scanner.Scan() {
		args = append(args, "--hostnames", scanner.Text())
	}

	setupCmd := exec.Command(adminPath, args...)
	output, err := setupCmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error: ", err)
	}

	fmt.Println(string(output))
}

func expandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, path[2:])
	}
	return filepath.Abs(path)
}

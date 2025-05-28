package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"

	"github.com/wan_jm/servlet/astinfo"
)

func main() {
	projectPath := flag.String("p", ".", "project root directory")
	flag.Parse()

	absPath, err := filepath.Abs(*projectPath)
	if err != nil {
		log.Fatalf("Error getting absolute path: %v", err)
	}

	project := &astinfo.Project{
		Path: absPath,
	}

	if err := project.Parse(); err != nil {
		log.Fatalf("Project parse failed: %v", err)
	}

	// TODO: Implement GenCode method
	fmt.Println("Project parsed successfully:")
	fmt.Printf("Module: %s\n", project.Module)
	fmt.Printf("Found %d packages\n", len(project.Packages))
}

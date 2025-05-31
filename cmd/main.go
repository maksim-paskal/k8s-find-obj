package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/maksim-paskal/k8s-find-obj/internal"
)

func main() {
	ctx := context.Background()

	application := internal.NewApplication()

	flag.StringVar(&application.Kubeconfig, "kubeconfig", os.Getenv("KUBECONFIG"), "Path to the kubeconfig file to use for CLI requests.")
	flag.StringVar(&application.WhereToSearch, "where", "*", "Where to run the application. Options: local, cluster")
	flag.StringVar(&application.WhatToSearch, "find", "", "What to search for.")
	flag.StringVar(&application.Namespace, "namespace", "", "Namespace to use for the search.")
	flag.StringVar(&application.Except, "except", "", "What to exclude from the search.")

	flag.Parse()

	if err := application.Validate(); err != nil {
		log.Fatal(err)
	}

	if err := application.Init(ctx); err != nil {
		log.Fatal(err)
	}

	if err := application.Run(ctx); err != nil {
		log.Fatal(err)
	}
}

package main

import "github.com/riad804/github-auth-manager/cmd"

var AppVersion = "dev"

func main() {
	cmd.Version = AppVersion
	cmd.Execute()
}

/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"github.com/sirupsen/logrus"
	"mydocker/cmd"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	cmd.Execute()
}

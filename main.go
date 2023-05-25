/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"github.com/wordstospend/goblackboard/board"
	"github.com/wordstospend/goblackboard/cmd"
)

func main() {
	_ = board.NewBlackBoard()
	cmd.Execute()

}

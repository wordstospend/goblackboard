/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wordstospend/goblackboard/board"
)

// fromJsonCmd represents the fromJson command
var fromJsonCmd = &cobra.Command{
	Use:   "fromJson",
	Short: "accept a json file as the initial blackboard",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fromJson()
		fmt.Println("fromJson called")

	},
}

func init() {
	rootCmd.AddCommand(fromJsonCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// fromJsonCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// fromJsonCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func fromJson() {
	fmt.Println("starting blackboard expermentation")
	a := board.Key{
		WatchKey: "a",
		TypeKey:  "string",
	}

	s := board.Key{
		WatchKey: "s",
		TypeKey:  "string",
	}

	testValue := []byte("ABC")

	A := board.Task{
		Consumer: func(resultChannel board.ResultChannel, argv ...board.Result) {
			resultChannel <- board.Result{Key: a, Value: testValue}

		},
		TriggerKeys: []board.Key{s},
	}

	blackBoard := board.NewBlackBoard()
	blackBoard.AddTask(A)

	blackBoard.Publish(s, []byte("unused"))

	results, err := blackBoard.Wait([]board.Key{a}, 1000)
	if err != nil {
		fmt.Printf("err %v\n", err)
	}
	fmt.Printf("result %v\n", results)

}

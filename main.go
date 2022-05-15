package main

import (
	"flag"
	"main/backend"
	"main/server"
	"os"

	"github.com/pterm/pterm"
)

func main() {

	// Parse the command line arguments
	if len(os.Args) < 2 {
		pterm.Error.Println("Please provide the CSV file with the question and answers as an argument")
		os.Exit(1)
	}
	csvFileArg := os.Args[1]

	//Extract the questions and answers
	backend.ExtractQuestionsAndAnswers(csvFileArg)

	// Parse the port number
	portPtr := flag.Int("port", 8080, "port to listen on")

	//Start the server on the specified port
	server.StartServer(*portPtr)

	//Print the banner
	pterm.DefaultCenter.Println(pterm.DefaultHeader.WithFullWidth().WithBackgroundStyle(pterm.NewStyle(pterm.BgLightRed)).WithMargin(10).Sprint("Athena: The Intelligent Q&A Search Engine"))
	pterm.Info.Println("(c)2022 by Akhil Datla")

}

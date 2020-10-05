package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/evantbyrne/terrarium/src"
)

func usage() {
	fmt.Println("Usage: terrarium [-expires] [-s3-bucket] [-s3-region] <command> [args]")
	fmt.Println("\nFlags:")
	flag.PrintDefaults()
	fmt.Println("\nCommands:")
	fmt.Println("  commit          Upload local state, then unlock remote state")
	fmt.Println("  download        Download remote state without locking")
	fmt.Println("  help            Prints helpful information about other commands")
	fmt.Println("  lock            Download and lock remote state")
	fmt.Println("  version         Prints the terrarium version")
}

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}

func main() {
	expires := flag.Uint64("expires", 600, "Maximum time in seconds lock the remote state.")
	s3Bucket := flag.String("s3-bucket", "", "S3 bucket.")
	s3Region := flag.String("s3-region", "", "S3 region.")

	flag.Usage = usage

	flag.Parse()

	args := flag.Args()
	if len(args) == 0 || (len(args) == 1 && args[0] == "help") {
		usage()
		os.Exit(1)
	}

	config := &src.Config{}
	if err := config.SetExpires(*expires); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	if err := config.SetS3Bucket(*s3Bucket); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	if err := config.SetS3Region(*s3Region); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	if err := config.SetS3AccessKey(os.Getenv("AWS_ACCESS_KEY")); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	if err := config.SetS3SecretKey(os.Getenv("AWS_SECRET_KEY")); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	os.Setenv("AWS_REGION", config.S3Region)

	var commands = map[string]src.Command{
		"lock": &src.CommandLock{},
	}
	command, ok := commands[args[0]]
	if !ok {
		fmt.Printf("Error: Invalid command '%s'.\n\n", args[0])
		usage()
		os.Exit(1)
	}

	if err := command.Run(config, args); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

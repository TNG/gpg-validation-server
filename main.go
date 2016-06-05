package main

import (
	"fmt"
	"github.com/TNG/gpg-validation-server/gpg"
	"github.com/TNG/gpg-validation-server/validator"
	"github.com/codegangsta/cli"
	"log"
	"os"
)

const (
	okExitCode    = 0
	errorExitCode = 1
)

func appAction(c *cli.Context) error {
	fmt.Println("Args", c.Args())
	fmt.Println("host", c.String("host"))
	// TODO #11 Start the servers
	return nil
}

func processMailAction(c *cli.Context) error {
	var err error
	var inputMail *os.File

	inputFilePath := c.String("file")

	if inputFilePath == "" {
		inputMail = os.Stdin
	} else {
		inputMail, err = os.Open(inputFilePath)
		if err != nil {
			return fmt.Errorf("Cannot open mail file '%s': %s", inputFilePath, err)
		}
		defer func() { _ = inputMail.Close() }()
	}

	privateKeyPath := c.String("private-key")
	if privateKeyPath == "" {
		return fmt.Errorf("Invalid private key file path: %s", privateKeyPath)
	}
	privateKeyInput, err := os.Open(privateKeyPath)
	if err != nil {
		return fmt.Errorf("Cannot open private key file '%s': %s", privateKeyPath, err)
	}
	defer func() { _ = privateKeyInput.Close() }()

	gpgUtil, err := gpg.NewGPG(privateKeyInput, c.String("passphrase"))
	if err != nil {
		return fmt.Errorf("Cannot initialize GPG: %s", err)
	}

	result, err := validator.HandleMail(inputMail, gpgUtil)
	if err != nil {
		return fmt.Errorf("Cannot handle mail: %s", err)
	}

	log.Printf("Mail has valid signature: %v.\n", result.IsSigned())

	return nil
}

func cliErrorHandler(action func(*cli.Context) error) func(*cli.Context) cli.ExitCoder {
	return func(c *cli.Context) cli.ExitCoder {
		if err := action(c); err != nil {
			return cli.NewExitError(fmt.Sprint("Error: ", err), errorExitCode)
		}
		return nil
	}
}

// RunApp starts the server with the provided arguments.
func RunApp(args []string) {
	app := cli.NewApp()
	app.Name = "GPG Validation Service"
	app.Usage = "Run a server that manages email verification and signs verified keys with the servers GPG key."

	app.Commands = []cli.Command{
		{
			Name:   "process-mail",
			Usage:  "process an incoming email",
			Action: cliErrorHandler(processMailAction),
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "file",
					Value: "./test/mails/signed_request_enigmail.eml",
					// TODO Handle missing value, use better default
					Usage: "`FILE_PATH` of the mail file, omit to read from stdin",
				},
				cli.StringFlag{
					Name:  "private-key",
					Value: "./test/keys/test-gpg-validation@server.local (0x87144E5E) sec.asc.gpg",
					// TODO Handle missing value, use better default
					Usage: "`PRIVATE_KEY_PATH` to the private gpg key of the server",
				},
				cli.StringFlag{
					Name:  "passphrase",
					Value: "validation",
					// TODO Handle missing value, use better default.
					Usage: "`PASSPHRASE` of the private key",
				},
			},
		},
	}

	app.Action = cliErrorHandler(appAction)
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "host",
			Value: "localhost",
			Usage: "`HOST` of the mail server",
		},
	}

	err := app.Run(args)
	if err != nil {
		cli.OsExiter(errorExitCode)
	} else {
		cli.OsExiter(okExitCode)
	}
}

func main() {
	RunApp(os.Args)
}

package main

import (
	"io/ioutil"
	"rpc"
	"flag"
	"fmt"
	"os"
)

var (
	server = flag.Bool("s", false, "run a server instead of a client")
	format = flag.String("f", "vim", "output format (currently only 'vim' is valid)")
)

func getSocketFilename() string {
	user := os.Getenv("USER")
	if user == "" {
		user = "all"
	}
	return fmt.Sprintf("%s/acrserver.%s", os.TempDir(), user)
}

func serverFunc() {
	socketfname := getSocketFilename()
	daemon = NewAutoCompletionDaemon(socketfname)
	defer os.Remove(socketfname)

	rpcremote := new(RPCRemote)
	rpc.Register(rpcremote)

	daemon.acr.Loop()
}

func Cmd_AutoComplete(c *rpc.Client) {
	file, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic(err.String())
	}
	apropos := flag.Args()[1]
	abbrs, words := Client_AutoComplete(c, file, apropos)
	if len(words) != len(abbrs) {
		panic("Lengths should match!")
	}

	fmt.Printf("[")
	for i := 0; i < len(words); i++ {
		fmt.Printf("{'word': '%s', 'abbr': '%s'}", words[i], abbrs[i])
		if i != len(words)-1 {
			fmt.Printf(", ")
		}
	}
	fmt.Printf("]")
}

func Cmd_Close(c *rpc.Client) {
	Client_Close(c, 0)
}

var cmds = map[string]func(*rpc.Client) {
	"autocomplete": Cmd_AutoComplete,
	"close": Cmd_Close,
}

func clientFunc() {
	// client

	client, err := rpc.Dial("unix", getSocketFilename())
	if err != nil {
		fmt.Printf("Failed to connect to the ACR server\n%s\n", err.String())
		return
	}
	defer client.Close()

	if len(flag.Args()) > 0 {
		cmds[flag.Args()[0]](client)
	}
}

func main() {
	flag.Parse()
	if *server {
		serverFunc()
	} else {
		clientFunc()
	}
}

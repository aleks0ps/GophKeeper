package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

// help
// register
// login
// list
// put
// get

type Client struct {
	URL string
}

type CmdType int

// Errors
var ErrEmptyLogin error = errors.New("Empty login")
var ErrEmptyPassword error = errors.New("Empty password")
var ErrNotEnoughArgs error = errors.New("Not enough arguments")

const (
	CmdUnknown CmdType = iota
	CmdHelp
	CmdRegister
	CmdLogin
	CmdList
)

const (
	SCmdUnknown  string = "unknown"
	SCmdHelp            = "help"
	SCmdRegister        = "register"
	SCmdLogin           = "login"
	SCmdList            = "list"
)

var commandTypes map[string]CmdType = map[string]CmdType{
	SCmdUnknown:  CmdUnknown,
	SCmdHelp:     CmdHelp,
	SCmdRegister: CmdRegister,
	SCmdLogin:    CmdLogin,
	SCmdList:     CmdList,
}

func getCmdType(cmd string) CmdType {
	t, ok := commandTypes[cmd]
	if !ok {
		return CmdUnknown
	}
	return t
}

type Cmd struct {
	Name    string
	Type    CmdType
	Options map[string]string
}

func readCmd() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	cmd, err := reader.ReadString('\n')
	return cmd, err
}

func parseCmd(ctx context.Context, cmd string) (*Cmd, error) {
	cmd = strings.TrimRight(cmd, "\r\n")
	cmd = strings.ToLower(cmd)
	argv := strings.Split(cmd, " ")
	switch getCmdType(argv[0]) {
	case CmdHelp:
		fmt.Println(SCmdHelp)
	case CmdRegister:
		opts := argv[1:]
		c := Cmd{Name: SCmdRegister, Type: CmdRegister}
		c.Options = make(map[string]string)
		for _, o := range opts {
			if strings.HasPrefix(o, "-login=") {
				opt := strings.Split(o, "=")
				// return if login is missing
				if strings.TrimSpace(opt[1]) == "" {
					log.Println(ErrEmptyLogin)
					return nil, ErrEmptyLogin
				}
				c.Options["login"] = opt[1]
			} else if strings.HasPrefix(o, "-pass=") {
				opt := strings.Split(o, "=")
				if strings.TrimSpace(opt[1]) == "" {
					log.Println(ErrEmptyPassword)
					return nil, ErrEmptyPassword
				}
				c.Options["password"] = opt[1]
			} else {
				fmt.Println(ErrNotEnoughArgs)
				return nil, ErrNotEnoughArgs
			}
		}
		return &c, nil
	case CmdLogin:
		fmt.Println(SCmdLogin)
	case CmdList:
		fmt.Println(SCmdList)
	default:
		fmt.Println(SCmdUnknown)
	}
	return nil, nil
}

func (c *Client) execCmd(ctx context.Context, cmd *Cmd) error {
	switch cmd.Type {
	case CmdRegister:
		scmd := fmt.Sprintf(`{ "login":"%s", "password":"%s"`, cmd.Options["login"], cmd.Options["pass"])
		payload, err := json.Marshal([]byte(scmd))
		if err != nil {
			log.Println(err)
			return err
		}
		buf := bytes.NewBuffer(payload)
		resp, err := http.Post(c.URL+"/register", "application/json", buf)
		if err != nil {
			log.Println(err)
			return err
		}
		fmt.Printf("%v\n", resp)
	}
	return nil
}

func main() {
	ctx := context.Background()
	var URL string = "http://localhost:8080"
	client := Client{URL: URL}
	for {
		fmt.Print("> ")
		cmd, _ := readCmd()
		c, err := parseCmd(ctx, cmd)
		if err != nil {
			log.Fatal("%v", err)
		}
		client.execCmd(ctx, c)
	}
}

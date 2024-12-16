package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strings"

	"golang.org/x/net/publicsuffix"
)

// help
// register
// login
// list
// put
// get

type Client struct {
	URL  string
	Http *http.Client
}

type CmdType int
type DataType int

type User struct {
	ID       string `json:"id,omitempty"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

// structure for passing data
type Data struct {
	Type    string `json:"type"`
	Payload []byte `json:"payload"`
}

var (
	ErrEmptyLogin       error = errors.New("Empty login")
	ErrEmptyPassword    error = errors.New("Empty password")
	ErrNotEnoughArgs    error = errors.New("Not enough arguments")
	ErrUnknownCommand   error = errors.New("Unknown command")
	ErrUnknownOption    error = errors.New("Unknown option")
	ErrEmptyCommand     error = errors.New("Empty command")
	ErrEmptyOption      error = errors.New("Empty option")
	ErrEmptySecretClass error = errors.New("Empty secret class")
)

const (
	DataUnknown DataType = iota
	DataPassword
	DataText
	DataBinary
	DataCard
)

const (
	SDataUnknown  string = "unknown"
	SDataPassword        = "password"
	SDataText            = "text"
	SDataBinary          = "binary"
	SDataCard            = "card"
)

const (
	CmdUnknown CmdType = iota
	CmdHelp
	CmdRegister
	CmdLogin
	CmdList
	CmdPut
)

const (
	SCmdUnknown  string = "unknown"
	SCmdHelp            = "help"
	SCmdRegister        = "register"
	SCmdLogin           = "login"
	SCmdList            = "list"
	SCmdPut             = "put"
)

var commandTypes map[string]CmdType = map[string]CmdType{
	SCmdUnknown:  CmdUnknown,
	SCmdHelp:     CmdHelp,
	SCmdRegister: CmdRegister,
	SCmdLogin:    CmdLogin,
	SCmdList:     CmdList,
	SCmdPut:      CmdPut,
}

var dataTypes map[string]DataType = map[string]DataType{
	SDataUnknown:  DataUnknown,
	SDataPassword: DataPassword,
	SDataText:     DataText,
	SDataBinary:   DataBinary,
	SDataCard:     DataCard,
}

func getCmdType(cmd string) CmdType {
	t, ok := commandTypes[cmd]
	if !ok {
		return CmdUnknown
	}
	return t
}

func getCmdSType(ctype CmdType) string {
	for st, t := range commandTypes {
		if t == ctype {
			return st
		}
	}
	return SCmdUnknown
}

func getDataType(d string) DataType {
	t, ok := dataTypes[d]
	if !ok {
		return DataUnknown
	}
	return t
}

func getSDataType(dtype DataType) string {
	for sd, t := range dataTypes {
		if t == dtype {
			return sd
		}
	}
	return SDataUnknown
}

type Cmd struct {
	Name    string
	Type    CmdType
	DType   DataType
	Options map[string]string
}

func readCmd() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	cmd, err := reader.ReadString('\n')
	return cmd, err
}

func parseCmd(ctx context.Context, cmd string) (*Cmd, error) {
	cmd = strings.TrimRight(cmd, "\r\n")
	if cmd == "" {
		return nil, ErrEmptyCommand
	}
	cmd = strings.ToLower(cmd)
	argv := strings.Split(cmd, " ")
	cmdType := getCmdType(argv[0])
	cmdSType := getCmdSType(cmdType)
	switch cmdType {
	case CmdHelp:
		fmt.Println(SCmdHelp)
	case CmdRegister, CmdLogin:
		opts := argv[1:]
		// register or login
		c := Cmd{Name: cmdSType, Type: cmdType}
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
				return nil, ErrNotEnoughArgs
			}
		}
		return &c, nil
	case CmdList:
		c := Cmd{Name: SCmdList, Type: CmdList}
		return &c, nil
	case CmdPut:
		c := Cmd{Name: SCmdPut, Type: CmdPut}
		if argv[1] == "pass" {
			opts := argv[2:]
			if len(opts) < 2 {
				log.Println(ErrNotEnoughArgs)
				return nil, ErrNotEnoughArgs
			}
			c.DType = DataPassword
			c.Options = make(map[string]string)
			for _, o := range opts {
				if strings.HasPrefix(o, "-name=") {
					opt := strings.Split(o, "=")
					if strings.TrimSpace(opt[1]) == "" {
						log.Println(ErrEmptyOption)
						return nil, ErrEmptyOption
					}
					c.Options["name"] = opt[1]
				} else if strings.HasPrefix(o, "-pass=") {
					opt := strings.Split(o, "=")
					if strings.TrimSpace(opt[1]) == "" {
						log.Println(ErrEmptyOption)
						return nil, ErrEmptyOption
					}
					c.Options["pass"] = opt[1]
				} else {
					log.Println(ErrUnknownOption)
					return nil, ErrUnknownOption
				}
			}
		} else if argv[1] == "file" {
			opts := argv[2:]
			if len(opts) < 2 {
				log.Println(ErrNotEnoughArgs)
				return nil, ErrNotEnoughArgs
			}
			c.DType = DataBinary
			c.Options = make(map[string]string)
			for _, o := range opts {
				if strings.HasPrefix(o, "-name=") {
					opt := strings.Split(o, "=")
					if strings.TrimSpace(opt[1]) == "" {
						log.Println(ErrEmptyOption)
						return nil, ErrEmptyOption
					}
					c.Options["name"] = opt[1]
				} else if strings.HasPrefix(o, "-path=") {
					opt := strings.Split(o, "=")
					if strings.TrimSpace(opt[1]) == "" {
						log.Println(ErrEmptyOption)
						return nil, ErrEmptyOption
					}
					c.Options["path"] = opt[1]
				} else {
					log.Println(ErrUnknownOption)
					return nil, ErrUnknownOption
				}
			}
		}
		return &c, nil
	default:
		c := Cmd{Name: SCmdUnknown, Type: CmdUnknown}
		return &c, ErrUnknownCommand
	}
	return nil, nil
}

func (c *Client) execCmd(ctx context.Context, cmd *Cmd) error {
	switch cmd.Type {
	case CmdRegister, CmdLogin:
		user := User{ID: "", Login: cmd.Options["login"], Password: cmd.Options["pass"]}
		payload, err := json.Marshal(&user)
		if err != nil {
			log.Println(err)
			return err
		}
		buf := bytes.NewBuffer(payload)
		URL := c.URL
		if cmd.Type == CmdRegister {
			URL = URL + "/register"
		} else if cmd.Type == CmdLogin {
			URL = URL + "/login"
		}
		resp, err := c.Http.Post(URL, "application/json", buf)
		if err != nil {
			log.Println(err)
			return err
		}
		log.Printf("%v\n", resp.Status)
	case CmdList:
		URL := c.URL + "/list"
		resp, err := c.Http.Post(URL, "application/json", nil)
		if err != nil {
			log.Println(err)
			return err
		}
		buf := bytes.Buffer{}
		_, err = buf.ReadFrom(resp.Body)
		if err != nil {
			log.Println(err)
			return err
		}
		var dataList []Data
		err = json.Unmarshal(buf.Bytes(), &dataList)
		if err != nil {
			log.Println(err)
			return err
		}
		// Collect data into groups
		var passList []Data
		var fileList []Data
		for _, d := range dataList {
			if d.Type == SDataPassword {
				passList = append(passList, d)
			} else if d.Type == SDataBinary {
				fileList = append(fileList, d)
			}
		}
		// Show data groups
		if len(passList) > 0 {
			fmt.Println(">> PASSWORDS")
			for _, p := range passList {
				fmt.Printf("\t%s\n", string(p.Payload))
			}
		}
		if len(fileList) > 0 {
			fmt.Println(">> FILES")
			for _, p := range fileList {
				fmt.Printf("\t%s\n", string(p.Payload))
			}
		}
	case CmdPut:
		URL := c.URL + "/put"
		switch cmd.DType {
		case DataPassword:
			type Password struct {
				Name     string `json:"name"`
				Password string `json:"password"`
			}
			data := Data{Type: getSDataType(DataPassword)}
			pass := Password{Name: cmd.Options["name"], Password: cmd.Options["pass"]}
			jsonPass, err := json.Marshal(pass)
			if err != nil {
				log.Println(err)
				return err
			}
			data.Payload = jsonPass
			payload, err := json.Marshal(data)
			if err != nil {
				log.Println(err)
				return err
			}
			buf := bytes.NewBuffer(payload)
			resp, err := c.Http.Post(URL, "application/json", buf)
			if err != nil {
				log.Println(err)
				return err
			}
			log.Printf("%v\n", resp.Status)
		case DataBinary:
			URL = URL + "/binary"
			fileName := cmd.Options["name"]
			filePath := cmd.Options["path"]
			// use pipe to load large files
			pr, pw := io.Pipe()
			writer := multipart.NewWriter(pw)
			contentType := writer.FormDataContentType()
			go func() {
				file, err := os.Open(filePath)
				if err != nil {
					log.Println(err)
					pw.CloseWithError(err)
					return
				}
				defer file.Close()
				part, err := writer.CreateFormFile("file", fileName)
				if err != nil {
					log.Println(err)
					pw.CloseWithError(err)
					return
				}
				_, err = io.CopyN(part, file, 4096)
				if err != nil {
					log.Println(err)
					pw.CloseWithError(err)
					return
				}
				pw.CloseWithError(writer.Close())
			}()
			// load data from pipe
			req, err := http.NewRequest("POST", URL, pr)
			if err != nil {
				log.Println(err)
				return err
			}
			req.Header.Set("Content-Type", contentType)
			resp, err := c.Http.Do(req)
			if err != nil {
				log.Println(err)
				return err
			}
			fmt.Printf("%v\n", resp.Status)
		}
		//fmt.Printf("%v\n", cmd)
	}
	return nil
}

func main() {
	ctx := context.Background()
	var URL string = "http://localhost:8080"
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		log.Fatal(err)
	}
	client := Client{URL: URL, Http: &http.Client{Jar: jar}}
	for {
		fmt.Print("> ")
		cmd, err := readCmd()
		if err != nil {
			log.Fatal("%v", err)
		}
		c, err := parseCmd(ctx, cmd)
		if err != nil {
			if errors.Is(err, ErrEmptyCommand) {
				continue
			}
			log.Println(err)
			continue
		}
		client.execCmd(ctx, c)
	}
}

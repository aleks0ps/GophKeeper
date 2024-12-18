package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strings"

	"github.com/aleks0ps/GophKeeper/cmd/client/config"
	"github.com/aleks0ps/GophKeeper/cmd/client/version"
	"github.com/aleks0ps/GophKeeper/internal/app/db"
	"golang.org/x/net/publicsuffix"
)

type Client struct {
	URL      string
	Http     *http.Client
	Download string
}

var (
	ErrEmptyLogin       error = errors.New("Empty login")
	ErrEmptyPassword    error = errors.New("Empty password")
	ErrNotEnoughArgs    error = errors.New("Not enough arguments")
	ErrUnknownCommand   error = errors.New("Unknown command")
	ErrUnknownOption    error = errors.New("Unknown option")
	ErrUnknownData      error = errors.New("Unknown data")
	ErrEmptyCommand     error = errors.New("Empty command")
	ErrEmptyOption      error = errors.New("Empty option")
	ErrEmptySecretClass error = errors.New("Empty secret class")
	ErrShouldNotReach   error = errors.New("Should not reach")
)

type CmdType int

const (
	CmdUnknown CmdType = iota
	CmdHelp
	CmdRegister
	CmdLogin
	CmdList
	CmdPut
	CmdGet
)

const (
	SCmdUnknown  string = "unknown"
	SCmdHelp            = "help"
	SCmdRegister        = "register"
	SCmdLogin           = "login"
	SCmdList            = "list"
	SCmdPut             = "put"
	SCmdGet             = "get"
)

var commandTypes map[string]CmdType = map[string]CmdType{
	SCmdUnknown:  CmdUnknown,
	SCmdHelp:     CmdHelp,
	SCmdRegister: CmdRegister,
	SCmdLogin:    CmdLogin,
	SCmdList:     CmdList,
	SCmdPut:      CmdPut,
	SCmdGet:      CmdGet,
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

type Cmd struct {
	Name    string
	Type    CmdType
	RType   db.RecordType
	Options map[string]string
}

func readCmd() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	cmd, err := reader.ReadString('\n')
	cmd = strings.TrimSpace(cmd)
	return cmd, err
}

func parseCmd(ctx context.Context, cmd string) (*Cmd, error) {
	if cmd == "" {
		return nil, ErrEmptyCommand
	}
	cmd = strings.ToLower(cmd)
	argv := strings.Split(cmd, " ")
	cmdType := getCmdType(argv[0])
	cmdSType := getCmdSType(cmdType)
	switch cmdType {
	case CmdHelp:
		c := Cmd{Name: SCmdHelp, Type: CmdHelp}
		return &c, nil
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
			c.RType = db.RecordPassword
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
			c.RType = db.RecordBinary
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
		} else if argv[1] == "card" {
			opts := argv[2:]
			if len(opts) < 5 {
				log.Println(ErrNotEnoughArgs)
				return nil, ErrNotEnoughArgs
			}
			c.RType = db.RecordCard
			c.Options = make(map[string]string)
			for _, o := range opts {
				if strings.HasPrefix(o, "-name=") {
					opt := strings.Split(o, "=")
					if strings.TrimSpace(opt[1]) == "" {
						log.Println(ErrEmptyOption)
						return nil, ErrEmptyOption
					}
					c.Options["name"] = opt[1]
				} else if strings.HasPrefix(o, "-number=") {
					opt := strings.Split(o, "=")
					if strings.TrimSpace(opt[1]) == "" {
						log.Println(ErrEmptyOption)
						return nil, ErrEmptyOption
					}
					c.Options["number"] = opt[1]
				} else if strings.HasPrefix(o, "-cvv=") {
					opt := strings.Split(o, "=")
					if strings.TrimSpace(opt[1]) == "" {
						log.Println(ErrEmptyOption)
						return nil, ErrEmptyOption
					}
					c.Options["cvv"] = opt[1]
				} else if strings.HasPrefix(o, "-month=") {
					opt := strings.Split(o, "=")
					if strings.TrimSpace(opt[1]) == "" {
						log.Println(ErrEmptyOption)
						return nil, ErrEmptyOption
					}
					c.Options["month"] = opt[1]
				} else if strings.HasPrefix(o, "-year=") {
					opt := strings.Split(o, "=")
					if strings.TrimSpace(opt[1]) == "" {
						log.Println(ErrEmptyOption)
						return nil, ErrEmptyOption
					}
					c.Options["year"] = opt[1]
				} else {
					log.Println(ErrUnknownOption)
					return nil, ErrUnknownOption
				}
			} // for
		} else if argv[1] == "text" {
			opts := argv[2:]
			if len(opts) < 2 {
				log.Println(ErrNotEnoughArgs)
				return nil, ErrNotEnoughArgs
			}
			c.RType = db.RecordText
			c.Options = make(map[string]string)
			for _, o := range opts {
				if strings.HasPrefix(o, "-name=") {
					opt := strings.Split(o, "=")
					if strings.TrimSpace(opt[1]) == "" {
						log.Println(ErrEmptyOption)
						return nil, ErrEmptyOption
					}
					c.Options["name"] = opt[1]
				} else if strings.HasPrefix(o, "-text") {
					// read from stdin
					reader := bufio.NewReader(os.Stdin)
					fmt.Print("txt> ")
					text, _ := reader.ReadString('\n')
					if strings.TrimSpace(text) == "" {
						log.Println(ErrEmptyOption)
						return nil, ErrEmptyOption
					}
					c.Options["text"] = text
				} else {
					log.Printf("ERR:putText: %v: %s\n", ErrUnknownOption, o)
					return nil, ErrUnknownOption
				}
			}
		} // text
		return &c, nil
	case CmdGet:
		c := Cmd{Name: SCmdGet, Type: CmdGet}
		if argv[1] == "pass" {
			c.RType = db.RecordPassword
		} else if argv[1] == "text" {
			c.RType = db.RecordText
		} else if argv[1] == "binary" || argv[1] == "file" {
			c.RType = db.RecordBinary
		} else if argv[1] == "card" {
			c.RType = db.RecordCard
		}
		opts := argv[2:]
		if len(opts) < 1 {
			log.Println(ErrNotEnoughArgs)
			return nil, ErrNotEnoughArgs
		}
		c.Options = make(map[string]string)
		for _, o := range opts {
			if strings.HasPrefix(o, "-name=") {
				opt := strings.Split(o, "=")
				if strings.TrimSpace(opt[1]) == "" {
					log.Println(ErrEmptyOption)
					return nil, ErrEmptyOption
				}
				c.Options["name"] = opt[1]
			} else {
				log.Printf("ERR:get: %v: %s\n", ErrUnknownOption, o)
				return nil, ErrUnknownOption
			}
		}
		return &c, nil
	default:
		c := Cmd{Name: SCmdUnknown, Type: CmdUnknown}
		return &c, ErrUnknownCommand
	}
	return nil, ErrShouldNotReach
}

func (c *Client) execCmd(ctx context.Context, cmd *Cmd) error {
	switch cmd.Type {
	case CmdHelp:
		fmt.Println("Help:")
		fmt.Printf("  register -login=user -pass=pass\n")
		fmt.Printf("  login -login=user -pass=pass\n")
		fmt.Printf("  list\n")
		fmt.Printf("  get pass -name=Some_name\n")
		fmt.Printf("  put card -name=my -number=4242424242424242 -cvv=111 -month=02 -year=2025\n")
		fmt.Printf("  put pass -name=wifi -pass=12345678\n")
		fmt.Printf("  put text -name=Nospace -text\n")
		fmt.Printf("  put file -name=filename -path=/path/to/filename\n")
	case CmdRegister, CmdLogin:
		user := db.User{ID: "", Login: cmd.Options["login"], Password: cmd.Options["pass"]}
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
		var dataList []db.Record
		err = json.Unmarshal(buf.Bytes(), &dataList)
		if err != nil {
			log.Println(err)
			return err
		}
		// Collect data into groups
		var passList []db.Record
		var fileList []db.Record
		var cardList []db.Record
		var textList []db.Record
		for _, d := range dataList {
			if d.Type == db.SRecordPassword {
				passList = append(passList, d)
			} else if d.Type == db.SRecordBinary {
				fileList = append(fileList, d)
			} else if d.Type == db.SRecordCard {
				cardList = append(cardList, d)
			} else if d.Type == db.SRecordText {
				textList = append(textList, d)
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
		if len(cardList) > 0 {
			fmt.Println(">> CARDS")
			for _, p := range cardList {
				fmt.Printf("\t%s\n", string(p.Payload))
			}
		}
		if len(textList) > 0 {
			fmt.Println(">> TEXT")
			for _, p := range textList {
				fmt.Printf("\t%s\n", string(p.Payload))
			}
		}
	case CmdPut:
		URL := c.URL + "/put"
		switch cmd.RType {
		case db.RecordPassword:
			data := db.Record{Type: db.GetSRecordType(db.RecordPassword)}
			pass := db.Password{Name: cmd.Options["name"], Password: cmd.Options["pass"]}
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
		case db.RecordBinary:
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
				for {
					_, err = io.CopyN(part, file, 4096)
					if err == io.ErrUnexpectedEOF || err == io.EOF {
						break
					}
					if err != nil {
						log.Println(err)
						pw.CloseWithError(err)
						return
					}
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
		case db.RecordCard:
			data := db.Record{Type: db.GetSRecordType(db.RecordCard)}
			card := db.Card{Name: cmd.Options["name"],
				Number: cmd.Options["number"],
				Cvv:    cmd.Options["cvv"],
				Month:  cmd.Options["month"],
				Year:   cmd.Options["year"]}
			jsonCard, err := json.Marshal(card)
			if err != nil {
				log.Println(err)
				return err
			}
			data.Payload = jsonCard
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
		case db.RecordText:
			data := db.Record{Type: db.GetSRecordType(db.RecordText)}
			text := db.Text{Name: cmd.Options["name"], Text: cmd.Options["text"]}
			jsonText, err := json.Marshal(text)
			if err != nil {
				log.Println(err)
				return err
			}
			data.Payload = jsonText
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
		} // cmd.DType
	case CmdGet:
		URL := c.URL + "/get"
		if cmd.RType == db.RecordUnknown {
			log.Println(ErrUnknownData)
			return ErrUnknownData
		}
		data := db.Record{Type: db.GetSRecordType(cmd.RType)}
		var resp *http.Response
		if cmd.RType == db.RecordPassword {
			pass := db.Password{Name: cmd.Options["name"]}
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
			resp, err = c.Http.Post(URL, "application/json", buf)
			if err != nil {
				log.Println(err)
				return err
			}
		} else if cmd.RType == db.RecordText {
			text := db.Text{Name: cmd.Options["name"]}
			jsonText, err := json.Marshal(text)
			if err != nil {
				log.Println(err)
				return err
			}
			data.Payload = jsonText
			payload, err := json.Marshal(data)
			if err != nil {
				log.Println(err)
				return err
			}
			buf := bytes.NewBuffer(payload)
			resp, err = c.Http.Post(URL, "application/json", buf)
			if err != nil {
				log.Println(err)
				return err
			}
		} else if cmd.RType == db.RecordCard {
			card := db.Card{Name: cmd.Options["name"]}
			jsonCard, err := json.Marshal(card)
			if err != nil {
				log.Println(err)
				return err
			}
			data.Payload = jsonCard
			payload, err := json.Marshal(data)
			if err != nil {
				log.Println(err)
				return err
			}
			buf := bytes.NewBuffer(payload)
			resp, err = c.Http.Post(URL, "application/json", buf)
			if err != nil {
				log.Println(err)
				return err
			}
		} else if cmd.RType == db.RecordBinary {
			binary := db.Binary{Name: cmd.Options["name"]}
			jsonBinary, err := json.Marshal(binary)
			if err != nil {
				log.Println(err)
				return err
			}
			data.Payload = jsonBinary
			payload, err := json.Marshal(data)
			if err != nil {
				log.Println(err)
				return err
			}
			buf := bytes.NewBuffer(payload)
			resp, err = c.Http.Post(URL, "application/json", buf)
			if err != nil {
				log.Println(err)
				return err
			}
		}
		if resp.Header.Get("Content-Type") == "application/json" {
			defer resp.Body.Close()
			buf := bytes.Buffer{}
			_, err := buf.ReadFrom(resp.Body)
			if err != nil {
				log.Println(err)
				return err
			}
			var rec db.Record
			err = json.Unmarshal(buf.Bytes(), &rec)
			if err != nil {
				log.Println(err)
				return err
			}
			if rec.Type == db.SRecordPassword {
				var pass db.Password
				err := json.Unmarshal(rec.Payload, &pass)
				if err != nil {
					log.Println(err)
					return err
				}
				fmt.Printf("OK >> %+v\n", pass)
			} else if rec.Type == db.SRecordText {
				var text db.Text
				err := json.Unmarshal(rec.Payload, &text)
				if err != nil {
					log.Println(err)
					return err
				}
				fmt.Printf("OK >> %+v\n", text)
			} else if rec.Type == db.SRecordCard {
				var card db.Card
				err := json.Unmarshal(rec.Payload, &card)
				if err != nil {
					log.Println(err)
					return err
				}
				fmt.Printf("OK >> %+v\n", card)
			}
		} else {
			_, params, _ := mime.ParseMediaType(resp.Header.Get("Content-Type"))
			mr := multipart.NewReader(resp.Body, params["boundary"])
			var fpath string
			var file *os.File
			for part, err := mr.NextPart(); err == nil; part, err = mr.NextPart() {
				fpath = c.Download + "/" + part.FileName()
				if _, err := os.Stat(fpath); errors.Is(err, os.ErrNotExist) {
					// Create file if not exists
					file, err = os.OpenFile(fpath, os.O_RDWR|os.O_CREATE, 0644)
					if err != nil {
						log.Fatal("ERR:get:Binary: ", err)
					}
					defer file.Close()
				} else {
					fpath = part.FileName()
					file, err = os.CreateTemp(c.Download, fpath)
					if err != nil {
						log.Fatal(err)
					}
					defer file.Close()
				}
				bytes, err := ioutil.ReadAll(part)
				if err != nil {
					log.Fatal(err)
				}
				if _, err := file.Write(bytes); err != nil {
					log.Fatal(err)
				}
			}
			log.Printf(" %s downloaded\n", fpath)
		}
		log.Printf("%v\n", resp.Status)
	}
	return nil
}

func main() {
	ctx := context.Background()
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%s %s %s\n", version.Version, version.Date, version.GoVersion)
	opts := config.ParseOptions()
	// create download dir
	err = os.MkdirAll(opts.Download, 0750)
	if err != nil && !errors.Is(err, os.ErrExist) {
		log.Fatal(err)
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := Client{URL: opts.URL, Http: &http.Client{Jar: jar, Transport: tr}, Download: opts.Download}
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

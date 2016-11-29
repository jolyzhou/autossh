package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

type serverInfo struct {
	Host   string `json:"host"`
	Port   int    `json:"port"`
	User   string `json:"user"`
	Passwd string `json:"passwd"`
}

type areaSer struct {
	Area string     `json:"area"`
	Info serverInfo `json:"info"`
}

func connect(user, password, host string, port int) (*ssh.Session, error) {
	var (
		auth         []ssh.AuthMethod
		addr         string
		clientConfig *ssh.ClientConfig
		client       *ssh.Client
		session      *ssh.Session
		err          error
	)
	// get auth method
	auth = make([]ssh.AuthMethod, 0)
	auth = append(auth, ssh.Password(password))

	clientConfig = &ssh.ClientConfig{
		User:    user,
		Auth:    auth,
		Timeout: 30 * time.Second,
	}

	// connet to ssh
	addr = fmt.Sprintf("%s:%d", host, port)

	if client, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
		return nil, err
	}

	// create session
	if session, err = client.NewSession(); err != nil {
		return nil, err
	}

	return session, nil
}

func main() {
	var (
		allInfo []areaSer
		num     int
	)
	b, err := ioutil.ReadFile("./config.json")
	if err != nil {
		fmt.Println("ReadFile: ", err.Error())
	}
	json.Unmarshal(b, &allInfo)
	fmt.Println("================================================")
	fmt.Println("===================AUTO SSH=====================")
	fmt.Println("=================SERVER LIST====================")
	for i, elem := range allInfo {
		fmt.Println(i, ":", elem.Area)
	}
	fmt.Println("================================================")
	fmt.Println("=========please input the number of area==========")
	fmt.Print("=========NUMBER=======> ")
	fmt.Scanf("%d", &num)

	if num < 0 || num > len(allInfo) {
		fmt.Println("Please select the right number.")
		os.Exit(1)
	}
	session, err := connect(allInfo[num].Info.User, allInfo[num].Info.Passwd, allInfo[num].Info.Host, allInfo[num].Info.Port)
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	fd := int(os.Stdin.Fd())
	oldState, err := terminal.MakeRaw(fd)
	if err != nil {
		panic(err)
	}
	defer terminal.Restore(fd, oldState)

	// excute command
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	termWidth, termHeight, err := terminal.GetSize(fd)
	if err != nil {
		panic(err)
	}

	// Set up terminal modes
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // enable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	// Request pseudo terminal
	if err := session.RequestPty("xterm-256color", termHeight, termWidth, modes); err != nil {
		log.Fatal(err)
	}

	// session.Run("top")
	err = session.Shell()
	if err != nil {
		fmt.Println("Shell error: ", err)
		return
	}
	err = session.Wait()
	if err != nil {
		fmt.Println("Wait error: ", err)
		return
	}
}

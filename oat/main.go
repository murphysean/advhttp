package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"regexp"
	"strings"

	"github.com/murphysean/advhttp"
)

const (
	VERSION = "1.0.0"
)

var (
	clients        = make(map[string]*Client)
	users          = make(map[string]*User)
	oatClient      = flag.String("c", "", "Select the client (by name) from the config to use for this exectution.")
	oatUser        = flag.String("u", "", "Select the user (by name) from the config to use for this execution.")
	username       = flag.String("username", "", "Specify a username that will be used to authenticate and obtain a user access token.")
	password       = flag.String("password", "", "Specify a password that will be used to authenticate and obtain a user access token.")
	printTokenInfo = flag.Bool("ti", false, "Print out the json tokeninfo rather than the token itself.")
	nonewline      = flag.Bool("n", false, "Output the token with (n)o new line")
)

type Client struct {
	id          string
	secret      string
	scope       []string
	tokenep     string
	tokeninfoep string
}

type User struct {
	username string
	password string
}

//I would like a simple cli that will fetch tokens for me
func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s v%s:\n", os.Args[0], VERSION)
		fmt.Fprintln(os.Stderr, "oat [-c <client-name>] [-u <user-name>]")
		fmt.Fprintln(os.Stderr, "\t[-username <username>] [-password <password>]")
		fmt.Fprintln(os.Stderr, "\t[-n] [-ti]")
		fmt.Fprintln(os.Stderr, "\t['help']")
		fmt.Fprintln(os.Stderr, "")
		flag.PrintDefaults()

		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "OAT (O Auth Token is a small cli program designed to quickly fetch ")
		fmt.Fprintln(os.Stderr, "oauth2 tokens from a server for api access. It uses a config file, ")
		fmt.Fprintln(os.Stderr, "'.oatconfig', stored in your home directory to select credentials ")
		fmt.Fprintln(os.Stderr, "to use. The format of the config file is the fairly standard .ini ")
		fmt.Fprintln(os.Stderr, "format.")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "It looks something like:")
		fmt.Fprintln(os.Stderr, "---")
		fmt.Fprintln(os.Stderr, "[clientA]")
		fmt.Fprintln(os.Stderr, "\tclient_id=clientA's clientId")
		fmt.Fprintln(os.Stderr, "\tclient_secret=clientA's secret")
		fmt.Fprintln(os.Stderr, "\ttoken_endpoint=oauth2 token endpoint for clientA")
		fmt.Fprintln(os.Stderr, "\ttoken_info_endpoint=token_info_endpoint (optional)")
		fmt.Fprintln(os.Stderr, "\tscope=clientA's scope ask")
		fmt.Fprintln(os.Stderr, "[userB]")
		fmt.Fprintln(os.Stderr, "\tusername=userB's username")
		fmt.Fprintln(os.Stderr, "\tpassword=userB's password")
		fmt.Fprintln(os.Stderr, "---")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Use the -c flag to select a client by it's ['name']")
		fmt.Fprintln(os.Stderr, "Use the -u flag to select a user by their ['name']")
		fmt.Fprintln(os.Stderr, "\tor alternatively use the -username and -password flags")
		fmt.Fprintln(os.Stderr, "oat will default to a random client if none is provided and use ")
		fmt.Fprintln(os.Stderr, "\tthe client_credentials grant_type to obtain a token on behalf of ")
		fmt.Fprintln(os.Stderr, "\tthe client")
		fmt.Fprintln(os.Stderr, "If a user is selected (or a username and password provided), oat will ")
		fmt.Fprintln(os.Stderr, "\tinstead use the password grant_type and obtain a token for that ")
		fmt.Fprintln(os.Stderr, "\tuser, with the client set as the audience.")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Using oat to help with oauth2 in curl:")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "curl \"http://a.b/oauth2/tokeninfo?access_token=`oat -c myclient -n`\"")
		fmt.Fprintln(os.Stderr, "curl -H \"Authoriation: Bearer `oat -c myclient -n`\" url")
		fmt.Fprintln(os.Stderr, "")
	}
	flag.Parse()
	log.SetFlags(0)
	//Parse the config file ~/.oatconfig
	config := make(map[string]string)
	//Start with the home directory config
	if usr, err := user.Current(); err == nil {
		if _, err := os.Stat(usr.HomeDir + "/.oatconfig"); !os.IsNotExist(err) {
			config, err = parseOatConfig(config, usr.HomeDir+"/.oatconfig")
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return
			}
		} else {
			fmt.Fprintln(os.Stderr, err)
			return
		}
	} else {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	for k, v := range config {
		//Get name
		name := k[:strings.LastIndex(k, ".")]
		attr := k[strings.LastIndex(k, ".")+1:]
		//For property put in users and clients
		switch attr {
		case "client_id":
			if _, ok := clients[name]; !ok {
				clients[name] = new(Client)
			}
			clients[name].id = v
		case "client_secret":
			if _, ok := clients[name]; !ok {
				clients[name] = new(Client)
			}
			clients[name].secret = v
		case "token_endpoint":
			if _, ok := clients[name]; !ok {
				clients[name] = new(Client)
			}
			clients[name].tokenep = v
		case "token_info_endpoint":
			if _, ok := clients[name]; !ok {
				clients[name] = new(Client)
			}
			clients[name].tokeninfoep = v
		case "scope":
			if _, ok := clients[name]; !ok {
				clients[name] = new(Client)
			}
			clients[name].scope = strings.Split(v, " ")

		case "username":
			if _, ok := users[name]; !ok {
				users[name] = new(User)
			}
			users[name].username = v
		case "password":
			if _, ok := users[name]; !ok {
				users[name] = new(User)
			}
			users[name].password = v
		}
	}
	var selectedClient *Client = nil
	var selectedUser *User = nil

	if *oatClient == "" {
		for _, v := range clients {
			selectedClient = v
		}
	} else {
		selectedClient = clients[*oatClient]
	}

	if selectedClient == nil {
		fmt.Fprintln(os.Stderr, "No client selected or present in config")
	}

	//If a command line -u is found, use that user, else...
	if *oatUser != "" {
		selectedUser = users[*oatUser]
	} else if *username != "" && *password != "" {
		//If a command line -username and -password is found use that user
		selectedUser = new(User)
		selectedUser.username = *username
		selectedUser.password = *password
	}

	var token string
	var err error
	if selectedUser == nil {
		token, _, err = advhttp.GetClientCredentialsToken(selectedClient.tokenep, selectedClient.id, selectedClient.secret, selectedClient.scope)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
	} else {
		token, _, _, err = advhttp.GetPasswordToken(selectedClient.tokenep, selectedClient.id, selectedClient.secret, selectedUser.username, selectedUser.password, selectedClient.scope)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
	}

	if *printTokenInfo {
		ti, err := advhttp.GetTokenInformation(selectedClient.tokeninfoep, token)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
		bytes, err := json.Marshal(&ti)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
		fmt.Println(string(bytes))
		return
	}

	if *nonewline {
		fmt.Print(token)
	} else {
		fmt.Println(token)
	}
}

func parseOatConfig(config map[string]string, path string) (c map[string]string, err error) {
	c = config
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	context := ""

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}
		//If there is a [] on the line, then we've got a new 'context'
		re := regexp.MustCompile(`\[(.*)\]`)
		if sm := re.FindStringSubmatch(line); len(sm) > 0 {
			context = sm[1]
			continue
		}

		parts := strings.Split(scanner.Text(), "=")
		if len(parts) == 2 {
			if context == "" {
				c[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			} else {
				c[context+"."+strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
	}

	if err = scanner.Err(); err != nil {
		return
	}

	return
}

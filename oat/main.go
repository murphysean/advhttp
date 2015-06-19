package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/murphysean/advhttp"
	flag "github.com/ogier/pflag"
	"net/url"
	"os"
	"os/user"
	"regexp"
	"strings"
)

const (
	VERSION = "1.0.0"
)

var (
	clients        = make(map[string]*Client)
	users          = make(map[string]*User)
	oatClient      = flag.StringP("client", "c", "", "Select the client (by name) from the config to use for this exectution.")
	oatUser        = flag.StringP("user", "u", "", "Select the user (by name) from the config to use for this execution.")
	username       = flag.StringP("username", "U", "", "Specify a username that will be used to authenticate and obtain a user access token.")
	password       = flag.StringP("password", "p", "", "Specify a password that will be used to authenticate and obtain a user access token.")
	printTokenInfo = flag.BoolP("tokeninfo", "i", false, "Print out the json tokeninfo rather than the token itself.")
	nonewline      = flag.BoolP("nonewline", "n", false, "Output the token with (n)o new line")
	verbose        = flag.BoolP("verbose", "v", false, "Print extra information to stderr")
)

type Client struct {
	Id          string
	Secret      string
	Scope       []string
	Tokenep     string
	Tokeninfoep string
}

type User struct {
	Username string
	Password string
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s v%s:\n", os.Args[0], VERSION)
		fmt.Fprintln(os.Stderr, "oat [--client <client-name>] [--user <user-name>]")
		fmt.Fprintln(os.Stderr, "\t[--username <username>] [--password <password>]")
		fmt.Fprintln(os.Stderr, "\t[--nonewline] [--tokeninfo]")
		fmt.Fprintln(os.Stderr, "\t[-c -u -U -p -n -i]")
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
		fmt.Fprintln(os.Stderr, "Use the -v flag to turn on verbose output")
		fmt.Fprintln(os.Stderr, "Use the -c flag to select a client by it's ['name']")
		fmt.Fprintln(os.Stderr, "Use the -u flag to select a user by their ['name']")
		fmt.Fprintln(os.Stderr, "\tor alternatively use the -U and -p flags")
		fmt.Fprintln(os.Stderr, "oat will error out if no client is provided. Otherwise it will use ")
		fmt.Fprintln(os.Stderr, "\tthe client_credentials grant_type to obtain a token on behalf of ")
		fmt.Fprintln(os.Stderr, "\tthe selected client")
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
	//Parse the config file ~/.oatconfig
	config := make(map[string]string)
	//Start with the home directory config
	if usr, err := user.Current(); err == nil {
		if _, err := os.Stat(usr.HomeDir + "/.oatconfig"); !os.IsNotExist(err) {
			config, err = parseOatConfig(config, usr.HomeDir+"/.oatconfig")
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		} else {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	} else {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
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
			clients[name].Id = v
		case "client_secret":
			if _, ok := clients[name]; !ok {
				clients[name] = new(Client)
			}
			clients[name].Secret = v
		case "token_endpoint":
			if _, ok := clients[name]; !ok {
				clients[name] = new(Client)
			}
			clients[name].Tokenep = v
		case "token_info_endpoint":
			if _, ok := clients[name]; !ok {
				clients[name] = new(Client)
			}
			clients[name].Tokeninfoep = v
		case "scope":
			if _, ok := clients[name]; !ok {
				clients[name] = new(Client)
			}
			clients[name].Scope = strings.Split(v, " ")

		case "username":
			if _, ok := users[name]; !ok {
				users[name] = new(User)
			}
			users[name].Username = v
		case "password":
			if _, ok := users[name]; !ok {
				users[name] = new(User)
			}
			users[name].Password = v
		}
	}

	var selectedClient *Client = nil
	var selectedUser *User = nil

	selectedClient = clients[*oatClient]

	if selectedClient == nil {
		fmt.Fprintln(os.Stderr, "No client selected or present in config")
		os.Exit(1)
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "Using Client: %v\n\tid: %v\n\tep: %v\n", *oatClient, selectedClient.Id, selectedClient.Tokenep)
		fmt.Fprintf(os.Stderr, "Using Scope: \n\t%v\n", strings.Join(selectedClient.Scope, "\n\t"))
	}

	//If a command line -u is found, use that user, else...
	if *oatUser != "" {
		selectedUser = users[*oatUser]
		if selectedUser == nil {
			fmt.Fprintln(os.Stderr, "The user selected is not present in config")
			os.Exit(1)
		}
	} else if *username != "" && *password != "" {
		//If a command line -username and -password is found use that user
		selectedUser = new(User)
		selectedUser.Username = *username
		selectedUser.Password = *password
		*oatUser = *username
	}

	if selectedUser != nil && *verbose {
		fmt.Fprintf(os.Stderr, "Using User:%v\tusername:%v\n", *oatUser, selectedUser.Username)
	}

	var token string
	var err error
	if selectedUser == nil {
		if *verbose {
			cURL, err := url.Parse(selectedClient.Tokenep)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			cURL.User = url.UserPassword(selectedClient.Id, selectedClient.Secret)
			fmt.Fprintf(os.Stderr, "curl \"%v\" -d 'grant_type=client_credentials' -d 'scope=%v'\n\n", cURL.String(), strings.Join(selectedClient.Scope, " "))
		}
		token, _, err = advhttp.GetClientCredentialsToken(selectedClient.Tokenep, selectedClient.Id, selectedClient.Secret, selectedClient.Scope)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	} else {
		if *verbose {
			cURL, err := url.Parse(selectedClient.Tokenep)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			cURL.User = url.UserPassword(selectedClient.Id, selectedClient.Secret)
			fmt.Fprintf(os.Stderr, "curl \"%v\" -d 'grant_type=password' -d 'username=%v' -d 'password=%v' -d 'scope=%v'\n\n",
				cURL.String(), selectedUser.Username, selectedUser.Password, strings.Join(selectedClient.Scope, " "))
		}
		token, _, _, err = advhttp.GetPasswordToken(selectedClient.Tokenep, selectedClient.Id, selectedClient.Secret, selectedUser.Username, selectedUser.Password, selectedClient.Scope)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	if *printTokenInfo {
		ti, err := advhttp.GetTokenInformation(selectedClient.Tokeninfoep, token)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		bytes, err := json.Marshal(&ti)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
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

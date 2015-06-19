OAT
===

OAT (short for OAuth Token) is a command line utility wrapper around
advhttp's oauth2Client library. It is designed to make working with
oauth2 simple. It currently supports the following oauth2 Grant
Types:

- Client Credentials
- Password

Getting Started
---

The first step will be to create your OAT config file. OAT reads this 
file to find clients and users based on the command line selectors.

	touch ~/.oatconfig

You will need at least a client in the file. The file format follows 
the ini file format. A client includes:

1. Required
 1. client\_id (string)
 1. client\_secret (string)
 1. token\_endpoint (url)
1. Optional
 1. token\_info\_endpoint (url)
 1. scope (A space seperated list of requested scopes)

You can also include a user in the oat config file. When requested
a user will be used in a `password` grant type. A user includes:

1. username (string)
1. password (string)

Example File contents:

	[myclient]
		client_id=clientid
		client_secret=clientsecret
		token_endpoint=https://www.example.com/oauth2/token
		token_info_endpoint=https://www.example.com/oauth2/tokeninfo
		scope=https://example.com/auth/scope1 https://example.com/auth/scope2
	[myuser]
		username=username
		password=password

Using OAT
---

Now that you have your oat file configured, you can use oat. There
is a help document included with oat, to use it type `oat -h`. OAT
has a number of flags that you can set:

1. -c or --client (string) index into ~/.oatconfig file
1. -u or --user (string) index into ~/.oatconfig file
1. -s or --scope (string seperated list) provide scope
1. -U or --username (string) provide username
1. -p or --password (string) provide password
1. -i or --tokeninfo (bool) return token info object
1. -n or --nonewline (bool) don't print newline after token
1. -v or --verbose (bool) turn on verbose output

Some Examples:
---

Our first example will get a new client\_credentials token on behalf
of the 'myclient' in the example file:

	oat -c myclient

Our next example will get a new password token on behalf of the 
'myclient' client and the 'myuser' user in the example file:

	oat -c myclient -u myuser

This next example will get a password token on behalf of the 'myclient'
client and then provide user credentials on the command line and also
override the default client scope with new scope on the command line:

	oat -c myclient -s 'https://example.com/auth/scope3 https://example.com/auth/scope4' -U user1 -p password

The final example will get a new token and utilize that token in a curl
command:

	curl https://api.example.com/cool/story/bro -H "Authorization: Bearer `oat -nc myclient`"

Notice in this example that we used the -n toggle to turn off printing
a newline and we used the backticks to execute the command and then
dump the response back into the input of the curl command.

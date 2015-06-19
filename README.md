advhttp
===

Advhttp is a go utility library to help out with some common
http needs that aren't in the standard library. The library
can be broken up into the following functionality categories.

- Handlers
- Logging
- OAuth2
- Cross Origin Resource Sharing
- Reverse Proxy

Handlers
---

The included handlers try to build a handler stack to alleviate
boilerplate code in go server applications. A common setup
might look like:

	ch := advhttp.NewDefaultCorsHandler(http.DefaultServeMux)
	ph := advhttp.NewPanicRecoveryHandler(ch)
	lh := advhttp.NewLoggingHandler(ph, os.Stdout)
	
	http.ListenAndServe(":http", lh)

This creates a request pipeline that will eventually end up using
the default serve mux provided in the http library. But you'll
have Common Log Format logs printed to std out, panic recovery, 
and cors headers to allow cross origin resouce sharing.

Logging
---

You can also bypass the handlers and get a little bit more out of
the library on your own. I provide a custom ResponseWriter that
tracks bytes written and status code so that you can have those
after you write. These are typically included in logs.

	trw := advhttp.NewResponseWriter(w)
	fmt.Fprintf(os.Stderr, "Status: %v\n", trw.Status())
	fmt.Fprintf(os.Stderr, "Bytes Written: %v\n", trw.Length())
	fmt.Fprint(os.Stdout, trw.LogCommonExtended(r))

OAuth2
---

The OAuth2 portion of the library is primarily for applications
which may need to communicate with other apis via oauth2. The
library will cache a token so that it may be re-used accross
multiple calls and reduce latency.

	tracker, err := advhttp.NewClientCredentialsTokenTracker(tokenEndpoint, tokenInfoEndpoint, client_id, client_secret, scope)
	token, err := tracker.GetToken() //Get a new token or a cached token
	token, err := tracker.GetNewToken() //Force getting a new token
	ti, err := tracker.GetTokenInfo() //Return an object with the result from the tokenInfoEndpoint

	//You can also just use the apis yourself
	token, tokenExpires, refreshToken, err := advhttp.GetPasswordToken(tokenEndpoint, client_id, client_secret, username, password, scope)
	token, tokenExpires, err := advhttp.GetRefreshToken(tokenEndpoint, client_id, client_secret, refresh_token, scope)
	token, tokenExpires, err := advhttp.GetClientCredentialsToken(tokenEndpoint, client_id, client_secret string, scope []string)
	ti, err := advhttp.GetTokenInformation(tokenInfoEndpoint, token string)

Cross Origin Resource Sharing
---

The cors library allows you to set up cors how you'd like, and then
serve based on that. advhttp comes with a DefaultCors object that
contains the default cors settings. `advhttp.DefaultCors`. A cors
object includes:

- AllowOrigin (string) Default "" meaning to mirror the Origin header
- AllowHeaders ([]string) Default advhttp.CorsDefaultAllowHeaders
- AllowMethods ([]string) Default advhttp.CorsDefaultAllowMethods
- ExposeHeaders ([]string) Default advhttp.CorsDefaultExposeHeaders
- MaxAge (int64) Default advhttp.CorsDefaultMaxAge
- AllowCredentials (bool) Default advhttp.CorsDefaultAllowCredentials

You can use the default cors handler:

	advhttp.ProcessCors(w,r)

Or you can create your own:

	cors := new(Cors)
	cors.AllowOrigin = "*"
	cors.AllowHeaders = []string{"Location", "Content-Type", "ETag", "Accept-Patch"}
	cors.AllowMethods = []string{"OPTIONS", "HEAD", "GET", "POST", "PUT", "PATCH", "DELETE"}
	cors.ExposeHeaders = []string{"Location", "Content-Type", "ETag", "Accept-Patch"}
	cors.MaxAge = int64(1728000)
	cors.AllowCredentials = true
	cors.ProcessCors(w,r)


Or you can use the included handler:

	ch := advhttp.NewDefaultCorsHandler(http.DefaultServeMux)
	ch.ServeHTTP(w,r)
	
Reverse Proxy
---

The reverse proxy adds to the httputils and allows you to reverse proxy
to different hosts, mutate the path on it's way there, (and also in the
Location header in responses). It is helpful in gateway projects.

	googleURL, _ := url.Parse("http://www.google.com/")
	downloads := advhttp.NewGatewayReverseProxy(googleURL, true, "/google/")
	http.Handle("/google/", someHandlerFunc)

In the above example, calls to the /google/ endpoint on your server 
`http://yourserver.com/google/` will result in your server then calling
google `http://www.google.com/` but stripping off the /google/ path 
from the original url. If the response from google is a redirect or
includes the `Location` header, then advhttp will rewrite the header
to include the original `/google/` path. This results in browser being
able to (almost) transparently use the other server through your 
server.

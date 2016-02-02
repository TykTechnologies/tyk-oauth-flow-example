Tyk OAuth Sample
================

This is a quick project that shows the Tyk OAuth request cycle from start to finish.

To try this project out:

1. In your Tyk Gateway, create an API and call it `oauth2`
2. Set the Access Method to "Oauth 2.0"
3. Select "Allowed Access Types: Authorization codes"
4. Select "Allowed Authorize Types: Token"
5. Set the login redirect for this API to be: `http://localhost:8000/login`
6. Take note of the API ID
7. Add an oauth client to it and set the redirect to be `http://localhost:8000/final`
8. Take note of the client ID
9. Create a policy that has access to this API, take not of the Policy ID

Now edit the `tmpl/index.html` file:

1. In the form elements, set the `redirect_uri` value to the one of your client
2. Set the `client_id` element to the value of your client ID

Now edit `config.go`:

1. Set the `APIlistenPath` to `oauth2` (or whatever the listen path is for your OAuth API)
2. Set `orgID` to be your Org ID (Go to users -> select your user, it is under RPC credentials)
3. Set `policyID` to be your policy ID
4. Set `GatewayHost` to be the host path to your gateway e.g. http://domain.com:port (note no trailing slash)
5. Set `AdminSecret` to your the secret in your `tyk.conf`

Now run the app:

	go run *.go

Then visit:

http://localhost:8000

If you've set everything up correctly, you should be taken throguh a full OAuth flow.

This app emulates two parties:

1. The requester (client)
2. The identity provider portal (your login page)

We make use of the Tyk REST API Authorization endpoint to complete the request cycle, you can see an API client in the `util.go` file.
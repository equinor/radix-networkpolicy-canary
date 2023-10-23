![allowradix](https://api.radix.equinor.com/api/v1/applications/radix-networkpolicy-canary/environments/allowradix/buildstatus)
# Radix Networkpolicy Canary
 
## Purpose


To have a canary test application which is used

- **Primarily** To verify that the Radix [egress rule feature](https://www.radix.equinor.com/guides/egress-config/) works as intended.
- **Secondarily** To verify that scheduling of [Radix jobs and Radix batch jobs](https://www.radix.equinor.com/guides/configure-jobs/) works as inteded.
- **Secondarily** To verify that the internal K8s DNS resolver in Radix works as inteded.

The endpoints of this application are queried in regular intervals by the [radix-cicd-canary application](https://github.com/equinor/radix-cicd-canary).

## Building and running using Docker

`docker build . -t networkpolicycanary && docker run --rm -e LISTENING_PORT=5000 networkpolicycanary`

NB! The `/testjobscheduler` and `/startjobbatch` will not work if the app is run outside Radix.

## Endpoints

- Server port configured by environment variables
- Log output to `stdout` and `stderr`
- /health - returns Status: 200
- /metrics - returns number of requests and errors
- /error - increases error count and returns HTTP 500 with an error
- /echo - returns the incomming request data including headers
- /testpublicdns - resolves any of 5 well known domains against Google and CloudFlare nameservers. Returns 200 OK if successful.
- /testinternaldns - resolves any of 5 well known domains against internal K8s nameservers. Returns 200 OK if successful.
- /testjobscheduler - creates job by querying Radix job scheduler. 200 OK if successful.
- /startjobbatch - creates job batch by querying Radix job scheduler. 200 OK if no error. NB! 200 OK does not guarantee job batch was scheduled successfuly. Success must be verified by querying K8s API and checking new job batch.
- /testexternalwebsite - queries any of 5 well known external websites. 200 OK if successful.
- /testradixsite - queries [the radix-canary-golang application](https://github.com/equinor/radix-canary-golang). 200 OK if successful.

In addition, the *oauthdenyall* environment also has standard OAuthv2 endpoints, including `/oauth2/callback`.

## Application environments

This Radix application has three different Radix environments where each environment has a different set of egress rules. 

* **egressrulestopublicdns:** Allows traffic to Google and CloudFlare public nameservers.
* **oauthdenyall:** Denies all traffic. Only allows traffic permitted by default rules.
* **allowradix:** Denies all traffic except to the Radix cluster which the app runs in.

The *oathdenyall* environment also has the [Radix OAuthv2 feature](https://www.radix.equinor.com/guides/authentication/#using-the-radix-oauth2-feature) enabled, but the other two environments do not. This table shows the expected result when the radix-cicd-canary app queries each endpoint in each app environment.

### Expected response from each endpoint in each app environment
|                      | egressrulestopublicdns | oauthdenyall | allowradix |
|----------------------|------------------------|--------------|------------|
| /testpublicdns       | <span style="color:green">200</span>                    | <span style="color:yellow">302</span>          | <span style="color:red">500</span>        |
| /testinternaldns     | <span style="color:green">200</span>                    | <span style="color:yellow">302</span>          | <span style="color:green">200</span>        |
| /testjobscheduler    | <span style="color:green">200</span>                    | <span style="color:yellow">302</span>          | <span style="color:green">200</span>        |
| /startjobbatch       | <span style="color:green">200</span>                    | <span style="color:yellow">302</span>          | <span style="color:green">200</span>        |
| /testexternalwebsite | <span style="color:red">500</span>                    | <span style="color:yellow">302</span>          | <span style="color:red">500</span>        |
| /testradixsite       | <span style="color:red">500</span>                    | <span style="color:yellow">302</span>          | <span style="color:green">200</span>        |
| /oauth2/callback?code=bullshitcode       | <span style="color:red">404</span>                    | <span style="color:red">500</span>          | <span style="color:red">404</span>        |

When the `/oauth2/callback?code=bullshitcode` endpoint in the *oauthdenyall* environment is queried, the radix-cicd-canary app expects a 500 response within 15 seconds. If the request takes longer, it is assumed that the radix-networkpolicy-canary app does not have network access to Microsoft's IDP endpoint, and the test is considered a failure.

### Which endpoints are queried by radix-cicd-canary in each app environment
Not all endpoints are queried in every app environment. This table shows the situation as of May 2022.

|  | egressrulestopublicdns | oauthdenyall | allowradix |
|---|---|---|---|
| /testpublicdns | Yes | No | No |
| /testinternaldns | Yes | No | No |
| /testjobscheduler | Yes | No | Yes |
| /startjobbatch | Yes | No | Yes |
| /testexternalwebsite | Yes | No | Yes |
| /testradixsite | Yes | No | Yes |
| /oauth2/callback?code=bullshitcode | No | Yes | No |


## Contributing

Read our [contributing guidelines](./CONTRIBUTING.md)

------------------

[Security notification](./SECURITY.md)
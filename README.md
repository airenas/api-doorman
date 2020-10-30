# api-doorman

![](https://github.com/airenas/api-doorman/workflows/Go/badge.svg) [![Coverage Status](https://coveralls.io/repos/github/airenas/api-doorman/badge.svg?branch=main)](https://coveralls.io/github/airenas/api-doorman?branch=main) [![Go Report Card](https://goreportcard.com/badge/github.com/airenas/api-doorman)](https://goreportcard.com/report/github.com/airenas/api-doorman) 

Simple proxy for traking API usage in DB and authenticating requests by keys.

The proxy is prepared to be used for Text-to-Spech application. We want to deny unlimited access to everyone.
Lets say we have API with a path *http://localhost:8002/private*. The API accepts JSON `{"text":"Some text to synthesize"}`. And we want to add quota for users based on count of characters in JSON's *text* field. The proxy can allow 1) access with some default quota to everyone (based on referrer's IP), 2) access with configured quota values for users with provided *key*.

## Demo

1. Go to [examples/docker-compose](examples/docker-compose)

1. Start a demo: `make start`

1. Test fake api by investigating *Makefile* and *docker-compose.yml*:
 
   ```bash
   make test-public
   make test-private
   make test-private-key
   ```

1. Add new key to DB: `make adm-add`
Expected result: 
    ```json
    {"key":"XK3JoSyC48cxgvvkpUF4", "manual":true,
    "validTo":"2030-11-24T11:07:00Z", "limit":500 ...}
    ```

1. Retrieve available keys from DB: `make adm-key-list`

1. Access private API: `make test-private-key key=<<created key>>` . Sample: `make test-private-key key=XK3JoSyC48cxgvvkpUF4`

**Note**: the proxy must be not exposed to the Internet directly! It is expected to work under some real proxy like: *nginx*, *traefik* or other. It uses *X-FORWARDED-FOR* header value to detect IP.

---

## License

Copyright © 2020, [Airenas Vaičiūnas](https://github.com/airenas).

Released under the [The 3-Clause BSD License](LICENSE).

---
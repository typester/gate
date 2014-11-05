# "gate" for your private resources

gate is a static file server and reverse proxy integrated with OAuth2 account authentication.

With gate, you can safely serve your private resources based on whether or not request user is a member of your company's Google Apps or GitHub organizations.

## Usage

1. Download [binary](https://github.com/typester/gate/releases) or `go get`
2. rename `config_sample.yml` to `config.yml`
3. edit `config.yml` to fit your environment
4. run `gate`

## Example config

```yaml
# address to bind
address: :9999

# # ssl keys (optional)
# ssl:
#   cert: ./ssl/ssl.cer
#   key: ./ssl/ssl.key

auth:
  session:
    # authentication key for cookie store
    key: secret123

  info:
    # oauth2 provider name (`google` or `github`)
    service: google
    # your app keys for the service
    client_id: your client id
    client_secret: your client secret
    # your app redirect_url for the service: if the service is Google, path is always "/oauth2callback"
    redirect_url: https://yourapp.example.com/oauth2callback

# # restrict user request. (optional)
# restrictions:
#   - yourdomain.com    # domain of your Google App (Google)
#   - example@gmail.com # specific email address (same as above)
#   - your_company_org  # organization name (GitHub)

# document root for static files
htdocs: ./

# proxy definitions
proxy:
  - path: /elasticsearch
    dest: http://127.0.0.1:9200
    strip_path: yes

  - path: /influxdb
    dest: http://127.0.0.1:8086
    strip_path: yes
```

## Authentication Strategy

gate now supports Google Apps and GitHub to authenticate users.

### Example config for Google

```yaml
auth:
  info:
    service: google
    client_id: your client id
    client_secret: your client secret
    redirect_url: https://yourapp.example.com/oauth2callback

# restrict user request. (optional)
restrictions:
  - yourdomain.com    # domain of your Google App
  - example@gmail.com # specific email address
```

### Example config for GitHub

Unlike the example of Google Apps above, if the `service` is GitHub, gate uses whether request user is a member of organization designated like below:

```yaml
auth:
  info:
    service: github
    client_id: your client id
    client_secret: your client secret
    redirect_url: https://yourapp.example.com/oauth2callback

# restrict user request. (optional)
restrictions:
  - foo_organization
  - bar_organization
```

#### github:e support

GitHub Enterprise is also supported. To authenticate via github enterprise, add api endpoint information to config like following:

```yaml
auth:
  info:
    service: github
    client_id: your client id
    client_secret: your client secret
    redirect_url: https://yourapp.example.com/oauth2callback
    endpoint: https://github.yourcompany.com
    api_endpoint: https://github.yourcompany.com/api
```

## Name Based Virtual Host

An example of "Name Based Viatual Host" setting.

```yaml
auth:
  session:
    # authentication key for cookie store
    key: secret123
    # domain of virtual hosts base host
    cookie_domain: gate.example.com

# proxy definitions
proxy:
  - path: /
    host: elasticsearch.gate.example.com
    dest: http://127.0.0.1:9200

  - path: /
    host: influxdb.gate.example.com
    dest: http://127.0.0.1:8086
```

## License

MIT

# "gate" for your private resources

gate is a static file server and reverse proxy integrated with google account authentications.

With gate, you can safely serve your private resources with your company google apps authenticaiton.

## Usage

1. rename `config_sample.yml` to `config.yml`
2. edit `config.yml` to fit your environment
3. run `gate`

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

  google:
    # your google app keys
    client_id: your client id
    client_secret: your client secret
    # your google app redirect_url: path is always "/oauth2callback"
    redirect_url: https://yourapp.example.com/oauth2callback

# # restrict domain. (optional)
# domain: yourdomain.com

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

## License

MIT

# GO HTTP RELAY

Microservice that allows you to carry the proxied traffic over an ordinary HTTP connection.

config.json can be placed inside the executable’s directory or /etc/go-http-relay/.

## Configuration example:
```yml
port: 7777
targetUrl: "https://api.telegram.org"
connectionTimeout: 5
proxy:
  url: "socks5://127.0.0.1:8999"
  username: "username"
  password: "password"
```
- port - the port http relay app will be listening on;
- targetUrl - request endpoint;
- proxy - proxy settings.
 
 ## Docker-Compose build example:
 ```
docker-compose build
docker-compose up -d
```

## Docker build example:
```
docker build -t relay .
```
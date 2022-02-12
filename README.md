
## Demo
https://www.youtube.com/watch?v=XKgUJN9BCp4&ab_channel=ChillMusic
##Setup telegram bot:
setup environment vartiable
- TG_TOKEN 

##Setup server:
### Setup this environment vartiable
- DB_ADDR
- DB_NAME
- DB_USER
- DB_PASS
- PORT

### Setup
 - setup GOOS for you os
 - rename .env.example to .env
 - rename .env.backend.exmaple to .env.backend

```shell
make build && make run
```

### example rdp file
```shell
auto connect:i:1
full address:s:ec2-3-120-129-136.eu-central-1.compute.amazonaws.com
username:s:Administrator
password somePassword
```
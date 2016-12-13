# GolangSystemInfo
This repo is a restful api to information system and execute command


## Routes
```
POST http://url/execute
GET  http://url/info
GET  http://url/ping
```

## Route : Execute

This route receive json parameters example 

```json
{
  "command" : "ls"
}

```
response: 

```
{
  "stdout": "bin\ngo.bash\nsrc\n",
  "stderr": ""
}
```

## Route : Info

response:
```json
{
  "os": {
    "arch": "amd64",
    "freemem": 188411904,
    "homedir": "/Users/jeffersonapolinario",
    "hostname": "jeffersonapolinario",
    "loadavg": [
      2.22,
      2.31,
      2.4
    ],
    "platform": "darwin",
    "release": "10.11.6",
    "tmpdir": "/var/folders/jc/__q22xmd3wxdtny_y4_bcvkm0000gq/T/",
    "totalmem": 12884901888,
    "uptime": 700432
  },
  "disk": [
    {
      "filesystem": "/dev/disk0s2",
      "size": 487546976,
      "used": 365385560,
      "available": 121905416,
      "capacity": 0.75,
      "amout": "/"
    },
    {
      "filesystem": "devfs",
      "size": 183,
      "used": 183,
      "available": 0,
      "capacity": 1,
      "amout": "/dev"
    },
    {
      "filesystem": "map -hosts",
      "size": 0,
      "used": 0,
      "available": 0,
      "capacity": 1,
      "amout": "/net"
    },
    {
      "filesystem": "map auto_home",
      "size": 0,
      "used": 0,
      "available": 0,
      "capacity": 1,
      "amout": "/home"
    }
  ]
}
```
## Package Manager
 * [Glide](https://github.com/Masterminds/glide)
 
## Dependecies
* [Echo](https://github.com/labstack/echo)
* [Gopsutil](https://github.com/shirou/gopsutil)


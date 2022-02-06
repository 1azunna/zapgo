# Zapgo
ZAP cli package for Dynamic Application Security Testing in CI/CD

## Why this project?
This package was created to make it easy for developers to perform dynamic application security testing with [OWASP ZAP](https://www.zaproxy.org/), making use of ZAP's automation framework.

## What are the benefits of usng zapgo?
- Run zap tests without worrying about zap setup commands. Focus only on your automation yaml file.
- Support for proxying postman collections through zap.
- Filter alerts by risk and confidence to reduce false positives.

## Installation

### Requirements

- Ensure docker is installed. if using in CI, ensure there is support for docker in docker.
- Internet connectivity to pull zap2docker and/or postman/newman docker images.  

### Using Go
Ensure $GOPATH/bin is added to the $PATH
```bash
go get github.com/1azunna/zapgo
```
### Github Releases

## Usage

```bash
Usage:
  zapgo [OPTIONS] <clean | init | run>

Application Options:
  -v, --verbose                 Show verbose output
      --release=[stable|weekly] The image release version (default: weekly)
      --port=                   Initialize ZAP with a custom port. (default: 8080)
  -p, --pull                    Pull the latest ZAP image from dockerhub
      --opts=                   Additional ZAP command line options to use when initializing ZAP

Help Options:
  -h, --help                    Show this help message

Available commands:
  clean  Clean Zapgo
  init   Initialize ZAP
  run    Run ZAP scan

[init command options]
    -n, --networkOnly       Create the zapgo-network without initializing the ZAP container.

[run command options]
        --file=                                  ZAP Automation framework config file. Automation file file must be placed within the current working directory..
        --collection=                            Postman collection file or url to run.
        --environment=                           Postman environment file or url to use with postman collection
        --policy=                                Import custom zap scan policy. Policy file must be placed within the current working directory.
    -c, --clean                                  Remove any existing zapgo containers and initialize ZAP.
        --confidence=[Low|Medium|High|Confirmed] Display alerts with confidence filter set to either Low, Medium, High or Confirmed. (default: Medium)
        --risk=[Low|Medium|High|Informational]   Display alerts with risk filter set to either Informational, Low, Medium, High. (default: Low)
        --fail=[Low|Medium|High]                 Set exit status to fail on a certain risk level. Allowed Risk levels are Low|Medium|High.


```

### Init
Use the init command to create the zapgo docker network and the zap container.  
Features of ZAP container:

- Name: zapgo-container
- Network: zapgo-network
- Hostname: zap
- Mount Location: Current working directory to /zap/wrk on the container.

If running tests against your docker application on the same host, ensure that the application is using the zapgo network. Run `zapgo init` or `zapgo init -n` before starting your application.
```bash
zapgo init
docker run -p 80:80 --network zapgo-network mywebapp
```
This allows tthe zap container to be aple to reach the docker application.

#### ZAP Startup Options
You can pass aditional zap command line options with Init or Run commands which will be useful for importing scripts. See available command line options [here](https://www.zaproxy.org/docs/desktop/cmdline/)

### Run
Use the run command to start a zap scan with zap's automation framework. The automation framework file can be specified using `--file=path/to/automation.yaml`. The file must be placed within the working directory or in a subfolder in the working directory.
```bash
zapgo init
zapgo run --file=automation.yaml
```
You can also initialize zap with run by passing the `-c` or `--clean` flag.
```bash
zapgo run -c --file=automation.yaml
```

### Run with Postman Colections
You can proxy postman requests through zap by using the `--collection` flag.
```bash
zapgo run -c --file=automation.yaml --collection=collection.json --environment=environment.json 
```

### Alert filtering
Filtering for alerts by risk and confidence, to reduce false positives is also a possibility. You can also set the exit status to 1 if issues are detected eg. `--risk=High --fail=Medium`
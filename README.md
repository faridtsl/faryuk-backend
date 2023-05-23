# FaRyuk

FaRyuk is a reconnaissance automation tool, and more when configured with custom docker images.

## Instalation

### Prerequisite

#### Database

Install a mongodb instance using your prefered method (native, docker, vm...)

#### Dependencies

Dependencies are normally installed automatically when running the program.

#### Screenshots
To be able to use the screenshot functionality, you will need to install the latest version of google-chrome (only chrome headless will be used). To install google-chrome :

```bash
# Download the package
wget https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb
# Install the package
sudo apt install ./google-chrome-stable_current_amd64.deb
```

#### Docker integration

The user you use to launch the server should have access to ` "/var/run/docker.sock:/var/run/docker.sock"` and should be in `docker` group.

## Running

### Config file

An example configuration file is provided with the repository.

### Ressources

To add port list files, wordlists and DNS wordlists, you should create this directory tree :
```bash
 mkdir ./ressources ./ressources/dirs ./ressources/ports  ./ressources/subdomains
```
 then add lis of ports to scan in a  `txt` file under `ressources/ports` and wirdlist (that will be detected by the server) under dirs.  
 
 **Example**   
 ports.txt
 ```txt
 80
 445
 8080
 ```
wordlist.txt
  ```txt
/robots.txt
/admin
 ```

### Run
```bash
go run main.go serve
```

## Disclaimer

Although FaRyuk is a security testing tool, it started as a script and comes with no garantee of its own security.

Please do not deploy it in a non controlled/hostile environnement.

Pull requests are welcome !

FROM golang:1.19-bullseye

RUN apt-get autoclean
RUN apt-get clean
RUN apt-get update

RUN apt-get install -y wget

RUN wget https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb
RUN apt-get install -y ./google-chrome-stable_current_amd64.deb

COPY . ${GOPATH}/src/faryuk
WORKDIR ${GOPATH}/src/faryuk

RUN go mod download
RUN go build -o ./faryuk -buildvcs=false
RUN mkdir -p ./ressources ./ressources/dirs ./ressources/ports  ./ressources/subdomains

# Set entrypoint and working directory
ENTRYPOINT ["./faryuk", "serve"]

# Indicate we want to expose ports 80 and 443
# EXPOSE 80/tcp 443/tcp


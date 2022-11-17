FROM golang:1.18-buster

RUN apt-get update && \
    apt-get install -y exiftool

RUN curl -sL https://aka.ms/InstallAzureCLIDeb | bash

COPY main.go ./

RUN curl -fsSL https://apt.releases.hashicorp.com/gpg | apt-key add -
RUN apt-get install -y software-properties-common
RUN apt-add-repository "deb [arch=amd64] https://apt.releases.hashicorp.com $(lsb_release -cs) main"
RUN apt-get update && apt-get install -y terraform

ENTRYPOINT ["go", "run" "main.go"]
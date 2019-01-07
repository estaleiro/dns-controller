# DNS CONTROLLER

Kubernetes controller for DNS.

References:

https://github.com/kubernetes/sample-controller

https://medium.com/@trstringer/create-kubernetes-controllers-for-core-and-custom-resources-62fc35ad64a3

https://medium.com/@cloudark/kubernetes-custom-controllers-b6c7d0668fdf

## Contributing

Go version: 1.11.4

1. Sync with git

```
mkdir /your_directory/dns-controller

export GOPATH=/your_directory/dns-controller/

mkdir $GOPATH/src/github.com/estaleiro/dns-controller/

cd $GOPATH/src/github.com/estaleiro/dns-controller/

git init

git remote add origin https://github.com/estaleiro/dns-controller.git

git config --global user.email "email@email.com"

git config --global user.name "usergithub"

git pull origin master
```

3. Build

```
export GO111MODULE=on

go mod tidy

go build -o dns-controller .
```

4. Using code generator

```
unset GO111MODULE

go get -u k8s.io/code-generator/...

cd $GOPATH/src/k8s.io/code-generator/

./generate-groups.sh all "github.com/estaleiro/dns-controller/pkg/client" "github.com/estaleiro/dns-controller/pkg/apis" "zone:v1"
```


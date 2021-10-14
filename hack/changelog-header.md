### Linux

```shell
curl -L https://github.com/{{.GitOwner}}/jx-context/releases/download/v{{.Version}}/jx-context-linux-amd64.tar.gz | tar xzv 
sudo mv jx-context /usr/local/bin
```

### macOS

```shell
curl -L  https://github.com/{{.GitOwner}}/jx-context/releases/download/v{{.Version}}/jx-context-darwin-amd64.tar.gz | tar xzv
sudo mv jx-context /usr/local/bin
```


### Linux

```shell
cd /tmp
curl -L https://github.com/{{.GitOwner}}/jx-context/releases/download/v{{.Version}}/jx-context_{{.Version}}_linux_amd64.tar.gz | tar xzv 
mv jx-context ~/.jx3/plugins/bin/jx-context-{{.Version}}
```

### macOS

```shell
cd /tmp
curl -L  https://github.com/{{.GitOwner}}/jx-context/releases/download/v{{.Version}}/jx-context_{{.Version}}_darwin_amd64.tar.gz | tar xzv
mv jx-context ~/.jx3/plugins/bin/jx-context-{{.Version}}
```


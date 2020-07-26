Simple tool to replace property placeholders like ${xxx} in any text file

# Install
```sh
go get jaspeen/propsubst
```
or download binary from releases page

# Usage
```sh
propsubst -f file.properties somefile.xml > target.xml
```

with inplace substitution
```sh
propsubst -f file.properties -i somefile.xml
```

with environment properties as additional properties source
```sh
propsubst -f file.properties -p "`env`" -i somefile.xml
```


# Salsa Demo

Please note that the `.salsarc` as present in this directory is actually the one
that should act as the user-specific configuration file since it contains the
HTTP user credentials and projects secrets. That should not be saved in any
source code repository.

### Publish

```bash
$ BUILD_NUMBER=92 BRANCH=master salsa -v -dry publish -tag x64 build/x64
Reading package.json ...
Reading $FILTERED/.salsarc ...
Reading .salsarc ...
Packing artifacts
    build/x64/1.txt
    build/x64/2.txt
    build/x64/3.txt
    build/x64/4.txt
    build/x64/5.txt
    build/x64/6.txt
    build/x64/7.txt
    build/x64/8.txt
    build/x64/9.txt
Archive created
Executing PUT https://artifacts.example.com/foobar-DZGscqnCP2NFkl7DnE3f/master/foobar-x64-master-1.2.3.92.tar
Archive uploaded to

  https://artifacts.example.com/foobar-DZGscqnCP2NFkl7DnE3f/master/foobar-x64-master-1.2.3.92.tar

```

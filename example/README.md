# Salsa Demo

You can run `salsa -v -dry SUBCMD` in this directory to see how it works.

### Publish

## OUT OF DATE

```bash
$ BUILD_NUMBER=56 salsa -v -dry -username Pepa -password Zdepa publish
Packing artifacts
    artifacts/1.txt
    artifacts/2.txt
    artifacts/3.txt
    artifacts/4.txt
    artifacts/5.txt
    artifacts/6.txt
    artifacts/7.txt
    artifacts/8.txt
    artifacts/9.txt
Archive created
Uploading the archive
  (using Basic authentication)
Archive uploaded to http://localhost:56789/art/foobar-1.2.3.56.tar
```

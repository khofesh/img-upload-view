# update load and view image services

## selinux

```shell
sudo chgrp -R nogroup configs
sudo chcon -Rt svirt_sandbox_file_t configs/
```

## environment

```shell
export UPLOAD_DIR="./upload/"
```

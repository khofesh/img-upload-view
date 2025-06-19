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

## requests

### upload

```shell
curl -X POST http://localhost/api/upload \
  -F "image=@/path/to/your/image.jpg" \
  -H "Content-Type: multipart/form-data"
```

### get all images

```shell
curl -X GET http://localhost/api/image

# with limit and offset
curl -X GET "http://localhost/api/image?limit=5&offset=0"

# next page
curl -X GET "http://localhost/api/image?limit=5&offset=5"
```

### get image by ID

```shell
curl -X GET http://localhost/api/image/1
```

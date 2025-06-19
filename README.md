# update load and view image services

`compose-dev.yaml` is for dev

`compose.yaml` is for prod (let's say it's prod)

## selinux

```shell
sudo chgrp -R nogroup configs
sudo chcon -Rt svirt_sandbox_file_t configs/
```

## development

docker

```shell
# docker compose
docker compose -f compose-dev.yaml up -d

export UPLOAD_DIR="./upload/"

make run/api
```

requests

```shell
curl -X POST http://localhost:8080/upload \
  -F "image=@/path/to/your/image.jpg" \
  -H "Content-Type: multipart/form-data"

# get all images
curl -X GET http://localhost:8080/images

# with limit and offset
curl -X GET "http://localhost:8080/images?limit=5&offset=0"

# next page
curl -X GET "http://localhost:8080/images?limit=5&offset=5"

# get image by ID
curl -X GET http://localhost:8080/images/1
```

psql

```shell
psql "postgres://postgres:postgres@localhost:5432/app_db?sslmode=disable"
```

frontend

```shell
export VITE_API_URL=http://localhost:8080
cd web
npm run dev
```

## environment

```shell
export UPLOAD_DIR="./upload/"
```

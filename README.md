# update load and view image services

`compose.dev.yaml` is for dev

`compose.yaml` is for prod (let's say it's prod)

## selinux

```shell
sudo chgrp -R nogroup configs
sudo chcon -Rt svirt_sandbox_file_t configs/
```

## development

docker and API service (terminal 1)

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

frontend (terminal 2)

```shell
export VITE_API_URL=http://localhost:8080
cd web
npm run dev
```

## fake prod

```shell
docker compose -f compose.yaml up --build # if "localhost" cannot be accessed, wait a bit
docker compose -f compose.yaml up -d
```

## generate dummy JPEG

```shell
cd dummy-jpeg
python3 gen-dummy-jpeg.py
```

test it on the webpage

![error-more-than-10mb](./images/error-more-than-10mb.png)

# Image Resizer Service (IRS)

![Go Version](https://img.shields.io/github/go-mod/go-version/dimakropachev/image-resizer-service)
![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)

[🇷🇺 RU](README.ru.md)

## Description

**IRS** is a web service that receives an image and converts it into images with the dimensions specified in the configuration file. Image transformations are handled by the workers, the number of which is also indicated in the configuration file.

## Starting

1. Clone the repository with the command:
```
git clone https://github.com/DimaKropachev/Image-Resizer-Service
```
2. Next, run the **Docker Engine**. If it is not installed, you can download it [here](https://www.docker.com/products/docker-desktop/)
3. Assemble the image from the Dockerfile with the following command:
```
docker build -f./DockerFile -t irs .
```
4. Now run the container based on the assembled image using the command:
```
docker run -p 8088:8088 irs
```

## IPI documentation 

|Endpoint|Method| Description|
|--------|-----|---------|
|/api/v1/upload|POST|image upload
|/api/v1/status|GET|getting the status of a task that contains the desired image|
|/api/v1/download|GET|download the finished image|
|/api/v1/delete|DELETE|delete all tasks or a specific|
|/api/v1/worker|GET|get statistics for a specific worker|

### POST /api/v1/upload

Allows you to upload an image to the server for further processing.

- **Request format**: `multipart/form-data` with the `image` field (**JPEG** or **PNG** file)
- **Response**: Unique identifier `task-id`

### GET /api/v1/status?id={task-id}

Allows you to find out the current state of image processing.

- **Request format**: `id` - the task ID received during the upload
- **Response**: the `status` field with one of the values:
  - `pending` - the task is in the queue
  - `processing` - the worker processes the image
  - `done` - processing is completed successfully
  - `failed` - an error has occurred (there will be a description in the `error` field)

### GET /api/v1/download?id={task-id}&size={size}

Returns the finished image or information about its readiness.

- **Request format**:
  - `id` - task ID 
  - `size` - the desired size (name from the configuration file)
- **Response**: If the task is not completed, its status will be returned. If the task is completed, the server will send the file `image/jpeg`

### DELTE /api/v1/delete?id={task-id}

Allows you to delete one or all of the tasks.

- **Request format**:
  - If the `id` is passed, only the specified task is deleted
  - If the `id` is not specified, all existing tasks are deleted.

### GET /api/v1/worker?id={worker-id}

Returns information about the work of a specific worker who processes images.

- **Request format**: `id` is the numeric identifier of the worker
- **Response**: Worker statistics

## License

This project is distributed under the MIT license.
<br>
For more information, see the [LICENSE](LICENSE) file.
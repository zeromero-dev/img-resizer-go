# Image Processing AP

### This is a simple image processing API that allows you to upload an image and get it back with a specific quality.

curl -X POST -F "image=@/path/to/yourImage" http://localhost:8080/api/images;

### To get the image with a specific quality

curl -X GET "http://localhost:8080/api/images/{id}?quality={amount(25,50,75)}" --output /path/to/output.jpg;

# To run the API

`go run cmd/api/main.go`

# To run worker

`go run cmd/worker/main.go`

# Don't forget to install the dependencie

`libvips-dev`

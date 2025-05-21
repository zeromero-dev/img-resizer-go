curl -X POST -F "image=@/path/to/yourImage" http://localhost:8080/api/images;
curl -X GET "http://localhost:8080/api/images/{id}?quality={amount(25,50,75)}" --output /path/to/output.jpg;

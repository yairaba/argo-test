Generate 10 branch and PR
```
docker-compose run bench go run main.go -create -label merge
```
Delete all branch
```
docker-compose run bench go run main.go -delete
```
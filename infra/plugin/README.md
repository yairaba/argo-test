curl --location --request POST 'http://localhost:8080/update' --header 'Content-Type: application/json' --data-raw '{
    "repoName": "myRepo",
    "branchName": "myBranch",
    "serviceData": {
        "myService1": "myValue1",
        "myService2": "myValue2",
        "myService3": "myValue3"
    }

curl --location --request POST 'http://localhost:8080/api/v1/template.execute' --header 'Content-Type: application/json' --data-raw '{
    "repoName": "myRepo",
    "branchName": "myBranch"
}'

Filter Read :

```
curl -X POST -H "Content-Type: application/json" -d '{"repoName": "my-repo", "branchName": "my-branch-aaa"}' https://plugin.inulogic.binboum.eu/api/v1/template.execute
```
All read :

```
curl -X POST -H "Content-Type: application/json" -d '{"repoName": "my-repo"}' https://plugin.inulogic.binboum.eu/api/v1/template.execute
```

Insert

```
curl -X POST -H "Content-Type: application/json" -d '{"repoName": "my-repo", "branchName": "my-branch-aaa", "serviceData": {"tag1": "nginx-aaa", "tag2": "mariadb-aaa"}}' https://plugin.inulogic.binboum.eu/update
```

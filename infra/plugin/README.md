Filter Read :

```
curl -X POST -H "Content-Type: application/json" -d '{"repo": "my-repo", "branch": "my-branch-aaa"}' https://plugin.inulogic.binboum.eu/api/v1/getparams.execute
```
All read :

```
curl -X POST -H "Content-Type: application/json" -d '{"repo": "my-repo"}' https://plugin.inulogic.binboum.eu/api/v1/getparams.execute
```

Insert

```
curl -X POST -H "Content-Type: application/json" -d '{"repo": "my-repo", "branch": "my-branch-aaa", "serviceData": {"tag1": "nginx-aaa", "tag2": "mariadb-aaa"}}' https://plugin.inulogic.binboum.eu/update
```

```
curl -X POST -H "Content-Type: application/json" -d '{"repo": "my-repo", "branch": "my-branch-bbb", "serviceData": {"tag1": "nginx-bbb", "tag2": "mariadb-bbb"}}' https://plugin.inulogic.binboum.eu/update
```

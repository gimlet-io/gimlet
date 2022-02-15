
## Pushing an artifact

```
export GIMLET_SERVER=http://127.0.0.1:8888
export GIMLET_TOKEN=

gimlet artifact push -f fixtures/artifact.json
gimlet artifact list
```

```
gimlet release make \
  --env staging \
  --app myapp \
  --artifact laszlocph/gimletd-test-repo-0633a684-b0ad-4bc8-a912-b3e1a306d904
```



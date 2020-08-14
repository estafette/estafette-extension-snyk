# extensions/snyk

This Estafette extension checks with Snyk API whether the running pipeline repository has any known vulnerabilities

## Parameters

| Parameter         | Type     | Values |
| ----------------- | -------- | ------ |
| `param1`          | string   | Document what this parameter does |

## Usage

In order to use this extension in your `.estafette.yaml` manifest for the various supported actions use the following snippets:

```yaml
check-snyk-for-vulernabilities:
  image: extensions/snyk:stable
```
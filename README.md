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

## Configure credential injection in Estafette CI

In order for this extension to be able to communicate with the Snyk API an api key needs to be configured in the `estafette-ci-api` config as follows:

```yaml
credentials:
- name: 'snyk-api-key'
  type: 'snyk-api-token'
  token: '<api key for your snyk account>'
```

Note: to ensure the api key isn't visible in plain text in the configuration you can encrypt it in Estafette's admin > secrets section; make sure to double encrypt it, otherwise it will already be decrypted in the config.

And to make sure the extension receives this credential it has to be configured as a trusted extension:

```yaml
trustedImages:
- path: extensions/snyk
  injectedCredentialTypes:
  - snyk-api-token
```
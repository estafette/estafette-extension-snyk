# extensions/snyk

This Estafette extension checks whether the running pipeline repository has any known vulnerabilities leveraging the Snyk CLI.

## Parameters

| Parameter            | Description                                                 | Allowed values                   | Default value                                  |
| -------------------- | ----------------------------------------------------------- | -------------------------------- | ---------------------------------------------- |
| `allProjects`        | Auto-detect all projects in working directory.              | bool                             | `false`                                        |
| `debug`              | Output debug logs.                                          | bool                             | `false`                                        |
| `exclude`            | Indicate sub-directories to exclude.                        | comma separated string           |                                                |
| `failOn`             | Only fail when there are vulnerabilities that can be fixed. | `all`, `upgradable`, `patchable` | `all`                                          |
| `file`               | Sets a package file.                                        | string                           |                                                |
| `packagesFolder`     | Custom path to packages folder.                             | string                           |                                                |
| `projectName`        | Specify a custom Snyk project name.                         | string                           | `${ESTAFETTE_GIT_OWNER}/${ESTAFETTE_GIT_NAME}` |
| `severityThreshold`  | Only report vulnerabilities of provided level or higher.    | `low`, `medium`, `high`          | `high`                                         |

## Usage

In order to use this extension in your `.estafette.yaml` manifest for the various supported actions use the following snippets:

```yaml
snyk-check:
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
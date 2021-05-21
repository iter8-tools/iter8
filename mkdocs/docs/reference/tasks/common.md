---
template: main.html
---

# Common Tasks

## `common/exec`

### Overview

The `common/exec` task executes a command with any specified arguments. Arguments are specified using templates that are instantiated at runtime using values defined in the versionavailable at the time.

### Arguments

| Field name | Field type | Description | Required |
| ----- | ---- | ----------- | -------- |
| cmd | string | The command that should be executed | Yes |
| args | []string | A list of command line arguments that should be passed to `cmd`. | No |
| disableInterpolation | bool | Flag indicating whether or not to disable the substitution of values in any templated arguments. Default is `false`. | No |

### Requirements

None.

### Result

The command will be executed.

### Examples

The following example executes the bash commands passed as arguments to `/bin/bash`:

```yaml
- start:
  task: common/exec
    with:
    - cmd: /bin/bash
    - args:
      - -c
      - |
        echo hello world
        echo good bye
```

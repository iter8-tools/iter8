---
template: main.html
title: "Iter8 K Assert"
hide:
- toc
---

## iter8 k assert

Assert if experiment result satisfies the specified conditions

### Synopsis


Assert if experiment result satisfies the specified conditions. 
If assert conditions are satisfied, exit with code 0. 
Else, return with code 1.

```
iter8 k assert [flags]
```

### Examples

```

# assert that the experiment completed without failures, 
# and SLOs were satisfied
iter8 assert -c completed -c nofailure -c slos

# another way to write the above assertion
iter8 assert -c completed,nofailure,slos

# if the experiment involves multiple app versions, 
# SLOs can be asserted for individual versions
# for example, the following command asserts that
# SLOs are satisfied by version numbered 0
iter8 assert -c completed,nofailures,slosby=0

# timeouts are useful for an experiment that may be long running
# and may run in the background
iter8 assert -c completed,nofailures,slosby=0 -t 5s

# assert that the most recent experiment running in the Kubernetes context is complete
iter8 k assert -c completed

```

### Options

```
  -c, --condition(s); can specify multiple or separate conditions with commas; strings   completed | nofailure | slos | slosby=<version number>
  -e, --experiment-id string                                                             remote experiment identifier; if not specified, the most recent experiment is used
  -h, --help                                                                             help for assert
  -t, --timeout duration                                                                 timeout duration (e.g., 5s)
```

### Options inherited from parent commands

```
      --as string                      Username to impersonate for the operation
      --as-group stringArray           Group to impersonate for the operation, this flag can be repeated to specify multiple groups.
      --cache-dir string               Default cache directory (default "/Users/srinivasanparthasarathy/.kube/cache")
      --certificate-authority string   Path to a cert file for the certificate authority
      --client-certificate string      Path to a client certificate file for TLS
      --client-key string              Path to a client key file for TLS
      --cluster string                 The name of the kubeconfig cluster to use
      --context string                 The name of the kubeconfig context to use
      --insecure-skip-tls-verify       If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
      --kubeconfig string              Path to the kubeconfig file to use for CLI requests.
  -n, --namespace string               If present, the namespace scope for this CLI request
      --password string                Password for basic authentication to the API server
      --request-timeout string         The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests. (default "0")
  -s, --server string                  The address and port of the Kubernetes API server
      --tls-server-name string         Server name to use for server certificate validation. If it is not provided, the hostname used to contact the server is used
      --token string                   Bearer token for authentication to the API server
      --user string                    The name of the kubeconfig user to use
      --username string                Username for basic authentication to the API server
```

### SEE ALSO

* [iter8 k](iter8_k.md)	 - Work with experiments running in a Kubernetes cluster

###### Auto generated by spf13/cobra on 10-Dec-2021
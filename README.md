# echoserver

traefik plugin for echo diagnostic info

## example

```yaml
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  name: echoserver
  namespace: traefik
spec:
  plugin:
    echo:
      path: /cdn-cgi/trace
```

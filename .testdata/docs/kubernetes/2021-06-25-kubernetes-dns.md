# Kubernetes - DNS Resolution

You can follow this post with the [interactive scenario](https://www.katacoda.com/bluebrown/scenarios/kubernetes-dns) on Katacoda.

We are going to explore how DNS resolution in Kubernetes works. First we create a `namespace`, then we create a `pod` and `expose` it via `service`. Afterwards, a second `pod` is used to perform DNS queries against the Kubernetes DNS Resolver to get the IP address of the service.

## Namespace

We create a namespace that we will use to bind our resources.

```shell
kubectl create namespace dev
```

Next, we run a pod via imperative command.

## Running some Pod

```shell
kubectl run my-app --image nginx --namespace dev --port 80
```

## Describe the Pod

```shell
kubectl describe pod my-app -n dev
```

<details>

```shell
Name:         my-app
Namespace:    dev
Priority:     0
Node:         raspberrypi/192.168.1.41
Start Time:   Fri, 25 Jun 2021 01:21:16 +0100
Labels:       run=my-app
Annotations:  <none>
Status:       Running
IP:           10.42.0.126
IPs:
  IP:  10.42.0.126
Containers:
  my-app:
    Container ID:   containerd://086772d833ec67917a98ef43561d6f18779f086daa5b93a3390474a6aa707160
    Image:          nginx
    Image ID:       docker.io/library/nginx@sha256:47ae43cdfc7064d28800bc42e79a429540c7c80168e8c8952778c0d5af1c09db
    Port:           80/TCP
    Host Port:      0/TCP
    State:          Running
      Started:      Fri, 25 Jun 2021 01:21:20 +0100
    Ready:          True
    Restart Count:  0
    Environment:    <none>
    Mounts:
      /var/run/secrets/kubernetes.io/serviceaccount from kube-api-access-fnb62 (ro)
Conditions:
  Type              Status
  Initialized       True
  Ready             True
  ContainersReady   True
  PodScheduled      True
Volumes:
  kube-api-access-fnb62:
    Type:                    Projected (a volume that contains injected data from multiple sources)
    TokenExpirationSeconds:  3607
    ConfigMapName:           kube-root-ca.crt
    ConfigMapOptional:       <nil>
    DownwardAPI:             true
QoS Class:                   BestEffort
Node-Selectors:              <none>
Tolerations:                 node.kubernetes.io/not-ready:NoExecute op=Exists for 300s
                             node.kubernetes.io/unreachable:NoExecute op=Exists for 300s
Events:
  Type    Reason     Age   From               Message
  ----    ------     ----  ----               -------
  Normal  Scheduled  30s   default-scheduler  Successfully assigned dev/my-app to raspberrypi
  Normal  Pulling    29s   kubelet            Pulling image "nginx"
  Normal  Pulled     28s   kubelet            Successfully pulled image "nginx" in 1.369183373s
  Normal  Created    28s   kubelet            Created container my-app
  Normal  Started    27s   kubelet            Started container my-app
```

</details>

Note the `label` that has been set by Kubernetes. `run=my-app` By default, Kubernetes will set labels that match the resource name. For resources started from a `run` it will have the form `run=<resource-name>`.

Now we can `expose` the pod. This will create a service matching the pods label.

## Create Service

```shell
kubectl expose pod my-app --namespace dev
```

## Check The Service

```shell
kubectl describe service my-app -n dev
```

<details>

```shell
Name:              my-app
Namespace:         dev
Labels:            run=my-app
Annotations:       <none>
Selector:          run=my-app
Type:              ClusterIP
IP Family Policy:  SingleStack
IP Families:       IPv4
IP:                10.43.52.98
IPs:               10.43.52.98
Port:              <unset>  80/TCP
TargetPort:        80/TCP
Endpoints:         10.42.0.126:80
Session Affinity:  None
Events:            <none>
```

</details>

Note how the service `selector` is matching the label `run=my-app`. That means it will match the pod we have previously deployed.

Now we can deploy another pod from which we query the Kubernetes DNS Resolver.

## Run dnsutils Pod

We run this pod in interactive mode and attach stdin so that we can use nslookup and dig from within the container to query the Kubernetes DNS Resolver.

```shell
kubectl run dnsutils --namespace dev --image tutum/dnsutils -ti -- bash
```

## Making DNS Queries

nslookup resolves the service ok

```shell
nslookup my-app
```

<details>

```shell
Server:         10.43.0.10
Address:        10.43.0.10#53

Name:   my-app.dev.svc.cluster.local
Address: 10.43.52.98
```

</details>

But dig doesn't find the service, why?

```shell
dig my-app
```

<details>

```shell
; <<>> DiG 9.11.5-P4-5.1+deb10u5-Debian <<>> my-app
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NXDOMAIN, id: 51094
;; flags: qr aa rd ra; QUERY: 1, ANSWER: 0, AUTHORITY: 1, ADDITIONAL: 1

;; OPT PSEUDOSECTION:
; EDNS: version: 0, flags:; udp: 4096
; COOKIE: 4c23b7d697ed3587 (echoed)
;; QUESTION SECTION:
;my-app.                                IN      A

;; AUTHORITY SECTION:
.                       13      IN      SOA     a.root-servers.net. nstld.verisign-grs.com. 2021062402 1800 900 604800 86400

;; Query time: 0 msec
;; SERVER: 10.43.0.10#53(10.43.0.10)
;; WHEN: Fri Jun 25 01:41:30 UTC 2021
;; MSG SIZE  rcvd: 122
```

</details>

## /etc/resolv.conf

In order to understand why dig doesn't find the service, let's take a look at /etc/resolv.conf

```shell
cat /etc/resolv.conf
```

<details>

```shell
search dev.svc.cluster.local svc.cluster.local cluster.local
nameserver 10.43.0.10
options ndots:5
```

</details>

This file contains a line with the following format.

```shell
search <namespace>.svc.cluster.local svc.cluster.local cluster.local
```

That means, when providing an incomplete part of the fully qualified domain name (FQDN), this file can be used to complete the query. However, dig doesn't do it by default. We can use the `+search` flag in order to enable it.

```shell
dig +search my-app
```

<details>

```shell
; <<>> DiG 9.11.5-P4-5.1+deb10u5-Debian <<>> +search my-app
;; global options: +cmd
;; Got answer:
;; WARNING: .local is reserved for Multicast DNS
;; You are currently testing what happens when an mDNS query is leaked to DNS
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 39376
;; flags: qr aa rd; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 1
;; WARNING: recursion requested but not available

;; OPT PSEUDOSECTION:
; EDNS: version: 0, flags:; udp: 4096
; COOKIE: de26c4eaa4e53026 (echoed)
;; QUESTION SECTION:
;my-app.dev.svc.cluster.local.  IN      A

;; ANSWER SECTION:
my-app.dev.svc.cluster.local. 5 IN      A       10.43.52.98

;; Query time: 0 msec
;; SERVER: 10.43.0.10#53(10.43.0.10)
;; WHEN: Fri Jun 25 01:42:34 UTC 2021
;; MSG SIZE  rcvd: 113
```

</details>

Now the service-name has been correctly resolved.

We can get the same service without `+search` flag when using the FQDN. The `+short` flag isn't required, but it will reduce the output to only the IP address.

```shell
$ dig +short my-app.dev.svc.cluster.local
10.43.52.98
```

However, the benefit of using the `search` method it that queries will automatically resolve to resources within the same namespace. This can be useful to apply the same configuration to different environments, such as production and development.

Resources in different namespaces always need to be looked up by the FQDN.

The same way the search entry in `resolv.conf` completes the query with the default name space, it will complete any part of the `FQDN` from left to right. So in the below example, it will resolve to the local cluster.

```shell
$ dig +short +search my-app.dev
10.43.52.98
```

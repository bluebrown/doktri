# Advanced Entrypoint Techniques for Docker Container

Building good images can be challenging. We want to provide enough abstraction and flexibility for the image to be used in different scenarios without having to rebuild them in every case. We also want to make it easy for users to use the image.

## CMD

When building docker images, it is oftentimes enough to use normal commands to start the container.

```dockerfile
FROM alpine
CMD ["echo", "Hello, world!"]
```

## Entrypoint

Occasionally, an additional entry point is used. When written like this, it is mostly for convenience, but there are some elaborate use cases.

```dockerfile
FROM alpine
ENTRYPOINT ["echo"]
CMD ["Hello, world!"]
```

## Entrypoint Script

Sometimes, we really need to do more work or operate on environment variables, which can be tricky due to the difference in shell and exec syntax in the Dockerfile. In that case, a script commonly called `entrypoint.sh` or `docker-entrypoint.sh` is executed as entrypoint.

```dockerfile
FROM alpine
ENV WORKER_SLEEP=5
COPY entrypoint.sh /
RUN chmod +x  /entrypoint.sh
ENTRYPOINT [ "/entrypoint.sh" ]
```

```shell
#!/bin/sh
set -e
echo 'Work, Work!'
sleep $WORKER_SLEEP
echo 'Hello, World!'
```

The `set -e` tells the shell to abort on first error. You can also set the `-x` flag to see the *execution plan*.

## Entrypoint Script with CMD

It is possible to still take the command list as arguments and access them in the script with the standard shell variables. Here, the CMD is executed via `exec $@` at the end of the `entrypoint.sh`.

```dockerfile
FROM alpine
ENV WORKER_SLEEP=5
COPY entrypoint.sh /
RUN chmod +x  /entrypoint.sh
ENTRYPOINT [ "/entrypoint.sh" ]
CMD echo 'Hello, World!'
```

```shell
#!/bin/sh
set -e
echo 'Work, Work!'
sleep $WORKER_SLEEP
exec $@
```

`$@` stands for the complete argument list. While the individual positional arguments can be accessed by `$n` where `n` stands for the arguments position in the list. e.g. `$1`, `$2`.

Another benefit is that the final application will become PID 1.

> This script uses the exec Bash command so that the final running application becomes the container’s PID 1. This allows the application to receive any Unix signals sent to the container. For more, see the ENTRYPOINT reference.

## Preserving Entrypoint Behavior

The argument list can be used in different way. So it doesn't only work with `exec`. Below the arguments are used like in the early entry point example but only after performing the work.

```dockerfile
FROM alpine
ENV WORKER_SLEEP=5
COPY entrypoint.sh /
RUN chmod +x  /entrypoint.sh
ENTRYPOINT [ "/entrypoint.sh" ]
CMD  ["Hello, World!"]
```

```shell
#!/bin/sh
set -e
echo 'Work, Work!'
sleep $WORKER_SLEEP
echo "$@"
```

Note how instead of `exec $@` it is `echo $@` so anything that was passed to as `command` will get echoed. You can also use `exec echo "$@"` which would run the `echo` command with PID 1 again.

## Shifting the Arguments

When building software for other people to use, it is always a good idea to anticipate all sorts of *misuse*. With entrypoint-scripts, it can be a good idea to leverage `shift` to correct the users *mistakes*.

The `Dockerfile` stays the same, but now we are checking if the first argument is `echo` and if so we remove if by shifting the argument list. `Shift` removes the first argument from the left and shifts the indices of the remaining arguments to the left.

```shell
#!/bin/sh
set -e
echo 'Work, Work!'
sleep $WORKER_SLEEP
[[ "$1" == "echo" ]] && shift
echo "$@"
```

That way, the user can run the image both ways.

```shell
docker run my-image Hello, World! # or
docker run my-image echo Hello, World!
```

## Real World Example

You can look at various official docker repos. For example, the official [nginx docker image on GitHub](https://github.com/nginxinc/docker-nginx/blob/master/entrypoint/docker-entrypoint.sh).

<details>
<summary>Nginx Entrypoint</summary>

```shell
#!/bin/sh
# vim:sw=4:ts=4:et

set -e

if [ -z "${NGINX_ENTRYPOINT_QUIET_LOGS:-}" ]; then
    exec 3>&1
else
    exec 3>/dev/null
fi

if [ "$1" = "nginx" -o "$1" = "nginx-debug" ]; then
    if /usr/bin/find "/docker-entrypoint.d/" -mindepth 1 -maxdepth 1 -type f -print -quit 2>/dev/null | read v; then
        echo >&3 "$0: /docker-entrypoint.d/ is not empty, will attempt to perform configuration"

        echo >&3 "$0: Looking for shell scripts in /docker-entrypoint.d/"
        find "/docker-entrypoint.d/" -follow -type f -print | sort -V | while read -r f; do
            case "$f" in
                *.sh)
                    if [ -x "$f" ]; then
                        echo >&3 "$0: Launching $f";
                        "$f"
                    else
                        # warn on shell scripts without exec bit
                        echo >&3 "$0: Ignoring $f, not executable";
                    fi
                    ;;
                *) echo >&3 "$0: Ignoring $f";;
            esac
        done

        echo >&3 "$0: Configuration complete; ready for start up"
    else
        echo >&3 "$0: No files found in /docker-entrypoint.d/, skipping configuration"
    fi
fi

exec "$@"
```

</details>

Another interesting one is from mongo db. It is fairly complex, with almost 400 lines. You can see the [full script](https://github.com/docker-library/mongo/blob/master/docker-entrypoint.sh) on GitHub.

It contains, amongst various others, a function using the shift method described earlier.

<details>
<summary>Shift Function</summary>

```shell
# _mongod_hack_ensure_no_arg '--some-unwanted-arg' "$@"
# set -- "${mongodHackedArgs[@]}"
_mongod_hack_ensure_no_arg_val() {
 local ensureNoArg="$1"; shift
 mongodHackedArgs=()
 while [ "$#" -gt 0 ]; do
  local arg="$1"; shift
  case "$arg" in
   "$ensureNoArg")
    shift # also skip the value
    continue
    ;;
   "$ensureNoArg"=*)
    # value is already included
    continue
    ;;
  esac
  mongodHackedArgs+=( "$arg" )
 done
}
```

</details>

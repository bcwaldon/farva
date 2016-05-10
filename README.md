# farva

# Development Environment

Use `gvm` and `govendor`.

1. `gvm pkgset create farva`
    Use gvm to create a pkgset to isolate dependencies.
2. `gvm pkgset use farva`
    Use gvm to manage your gopath.
3. `gvm linkthis github.com/bcwaldon/farva`
    Link this repository into the pkgset created by gvm.
4. `cd "$(echo $GOPATH | cut -f 1 -d :)/src/github.com/bcwaldon/farva"`
    Quick hack to cd into the correct directory assuming that the `gvm pkgset
    use` command has set the first component of your gopath to the directory
    it should have created and that linkthis successfully created an alias.

This is the start to a Kubernetes Ingress Controller.
Start by building the code:

```
% ./build
Building bin/darwin_amd64/farva...done
Building bin/linux_amd64/farva...done
```

Now, assuming you've got a Kubernetes cluster with a single service, you can run farva:

```
% ./bin/darwin_amd64/farva --kubeconfig=<KUBECONFIG>

http {

    server default__nginx {
        listen 30190;
    }
    upstream default__nginx {

        server 10.1.2.5;  # nginx-aheok
        server 10.1.2.7;  # nginx-avop9
        server 10.1.2.6;  # nginx-bguu4
    }

}
```

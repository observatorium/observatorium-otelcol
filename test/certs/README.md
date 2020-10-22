# Generate the certs

The certificates for the tests are not generated during build time, in order to simplify the scripts and reduce the build time. 

In case the certs need to be generated again, the following commands can be used:
```console
$ cd test/certs/
$ cfssl genkey -initca ca-csr.json | cfssljson -bare ca
$ cfssl gencert -ca ca.pem -ca-key ca-key.pem cert-csr.json | cfssljson -bare cert
$ chmod a+r *pem
```

This requires [`cfssl`](https://github.com/cloudflare/cfssl).

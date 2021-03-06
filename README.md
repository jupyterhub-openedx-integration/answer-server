An authenticated https server used for serving jupyter notebooks.

This server will not work without two .pem cert files.

At the moment, these two files are manually generated by running a
[python script](https://techoverflow.net/2021/07/18/how-to-export-certificates-from-traefik-certificate-store/) against the acme.json file used by Traefik which was
generated by the let's encrypt registration process.

The filenames are currently hardcoded, but for no good reason.

```go
	keyFile := "./key.pem"
	certFile := "./certificate.pem"
```

To start the server it's easier to use the makefile like this.

`$ make serve`

Which invokes the server like this:

`$ ./auth-file-server -secret ${GRADER_SECRET} -consumer ${GRADER_CONSUMER}`

where the two GRADER_ environmental variables are sourced from another
file not included in the repository for security reasons. It is called
env-secret.bash and looks like

```bash
export GRADER_CONSUMER=...
export GRADER_SECRET=...
```

These two values already exist within the wider system.

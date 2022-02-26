# Tests

As per @ktrysmt's comment [here](https://github.com/ktrysmt/go-bitbucket/pull/122#issuecomment-758373984) tests are being changed so they can be run against both:
1. The actual Bitbucket API
1. The Swagger documentation from Bitbucket ([https://bitbucket.org/api/swagger.json](https://bitbucket.org/api/swagger.json)), using [Stoplight's Prism](https://stoplight.io/open-source/prism)

The latter will eventually be run as part of a CI process using [Github Actions](https://github.com/features/actions).

## Running tests locally against the Bitbucket API

Please refer to [../README.md](../README.md).

# Running test locally against Prism

Run in a shell terminal:
```
docker run --rm -it -p 4010:4010 stoplight/prism:3 mock -h 0.0.0.0 https://bitbucket.org/api/swagger.json
```

Then in another shell terminal session run:
```
make test/swagger
```

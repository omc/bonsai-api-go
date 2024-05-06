# Contributing

## Pre-commit

This project uses [pre-commit](https://pre-commit.com/) to lint and store 3rd-party dependency licenses.
Installation instructions are available on the [pre-commit](https://pre-commit.com/) website!

To verify your installation, run this project's pre-commit hooks against all files:

```shell
pre-commit run --all-files
```

### Go-licenses pre-commit hook

Windows users: Ensure that you have `C:\Program Files\Git\usr\bin` added
to your `PATH`!
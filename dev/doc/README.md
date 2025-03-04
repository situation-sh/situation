# Doc generator

This code aims to generate markdown files from module source files.

Basically, if `DoStuffModule` is defined in `modules/do_stuff.go`, the tool generates `docs/modules/do_stuff.md`.

We cannot select a single module: the tool treats all the modules at once and also updates `docs/modules/index.md`.

From the root directory you can run this tool through:

```shell
make modules-doc
```

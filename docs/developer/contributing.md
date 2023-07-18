# Contributing

When contributing to this project, please first discuss the change you wish to make via Github **issue**.
The issue should be enough documented to well understand the bug or the requested feature.

## License

All the code submitted in this project follows the [project LICENSE]({{ github_repo }}/blob/main/LICENSE.md).

## MR Process

1. Fork the project into your personal namespace (or group) on Github.
2. Create a feature branch in your fork with naming `<issue-id>-<lowercase-title-of-the-issue>`.
3. Make changes (code, docs...)
4. Ensure the documentation is updated (`docs/` folder)
5. Push the commits to your feature branch in your fork.
6. Submit a pull request (PR) to the main branch in the main Github project.

## Coding Style

_Taken from [moby/moby](https://github.com/moby/moby/blob/master/CONTRIBUTING.md#coding-style)_

Unless explicitly stated, we follow all coding guidelines from the Go
community. While some of these standards may seem arbitrary, they somehow seem to result in a solid, consistent codebase.

It is possible that the code base does not currently comply with these
guidelines. We are not looking for a massive PR that fixes this, since that
goes against the spirit of the guidelines. All new contributions should make a best effort to clean up and make the code base better than they left it.
Obviously, apply your best judgement. Remember, the goal here is to make the
code base easier for humans to navigate and understand. Always keep that in
mind when nudging others to comply.

If you are having trouble getting into the mood of idiomatic Go, we recommend
reading through [Effective Go](https://golang.org/doc/effective_go.html). The
[Go Blog](https://blog.golang.org) is also a great resource. Drinking the
kool-aid is a lot easier than going thirsty.

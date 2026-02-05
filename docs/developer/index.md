______________________________________________________________________

## title: Overview summary: From the developer's perspective

The Situation project aims to be enriched by the community, and [modules](../modules/index.md) are definitely a good starting point for developers to contribute.

Before detailing the internals of Situation, it is paramount to understand the overall spirit of the project.

**No user interaction**: it means that the agent must run without configuration, without integration, without dependency. In some cases, we obviously need some extra information. In this project, the developer should code enough logic to guess what is missing. For instance, if you want to detect a database, you need to guess what could be its listening port.

Fortunately, modules also provide data that could be useful for subsequent modules through the [store](store.md). So developers should well define their dependencies to ease the workflow of their module. Basically, we should avoid to do twice the same thing.

**Security**: yes it is hard to ensure at 100%. However, for this kind of project, we quickly feel like using `exec.Command` and other shortcuts that ease developers' job (but decrease security level). So, do not use `exec.Command` and do not use library that uses it. In nutshell, we should keep in mind that this agent is likely to run with root privileges on critical systems.

# Architecture

The agent is divided into several modules.

| Module | Job |
| ---------- | ---------------------------------------------------------------- |
| `backends` | all the possible outputs (like stdout, file, databases...) |
| `config` | central app configuration |
| `cmd` | agent entrypoint (it basically manages the run of the agent) |
| `models` | definition of the models that represent what could be discovered |
| `modules` | all the collectors |
| `store` | internal payload where all the retrieved information are stored |
| `utils` | extra helpers |

The overall architecture is quite classical for plugin-based tools. An orchestrator schedules and runs the available modules and all the collected data can be sent in the backend you want.

~{architecture}(architecture.json)

# Home, runs, health, dependencies, release and settings

The side bar has the views of the terminal interface, in its order: Home, Projects, Runs, Health, Dependencies, Release and Settings.

- **Home** says what needs you now. Each item has its buttons, and each button shows the smith command it runs.
- **Runs** opens on every run you start: its tasks by project, and the log of a task on a click.
- **Health** groups the checks of `smith doctor` as Failed, Warnings and Passed. **Fix** shows the review, applies it only after you agree, and checks again.
- **Dependencies** lists what has a newer release: the patch and minor releases first, then the major ones. **Update** asks first and shows the `smith deps up` command.
- **Release** shows what `smith release` would publish and the checks before a release. It publishes nothing until you confirm.
- **Settings** lists the settings of `smith.toml` with their `smith config set` command.

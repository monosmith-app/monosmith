# tasks.json and the keyboard

Tasks of type `smith` appear in **Run Task** and can be written in `tasks.json`:

```json
{ "type": "smith", "task": "test", "affected": true, "problemMatcher": ["$smith", "$smith-tsc", "$smith-eslint"] }
```

| Keys | Does |
| --- | --- |
| `Ctrl+Alt+M T` | Run Task |
| `Ctrl+Alt+M A` | Run Affected |
| `Ctrl+Alt+M R` | Re-run Last |
| `Ctrl+Alt+M C` | Cancel Run |
| `Enter` on a task | run it |

Settings: `monosmith.refresh.auto`, `monosmith.refresh.debounceMs`, `monosmith.run.parallel`, `monosmith.run.bail`, `monosmith.run.noCache`.

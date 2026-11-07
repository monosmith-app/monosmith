# lazydocker: install notes

## What was installed

- Version 0.25.2 (Windows, amd64), from the GitHub release `lazydocker_0.25.2_Windows_x86_64.zip`.
- Command: `gh release download -R jesseduffield/lazydocker -p "*Windows_x86_64.zip"`, then `unzip`.
- Location: `.tools/lazydocker/lazydocker.exe`. Windows is supported natively.

## Docker state

- At the start, `docker info` failed: the Docker Desktop pipe `dockerDesktopLinuxEngine` did not exist.
- A few minutes later the daemon answered. I did not start it. I do not know what started it.
- Because of this, I did **not** capture the "daemon not reachable" screen of lazydocker. That case is not verified.
- The daemon holds an unrelated stack (images `langgenius/dify-*`, containers `docker-*`). I did not act on those containers.
  Some screenshots show their image and volume names in the lower panels.
- During one run a container `tuires-control-plane` appeared in the list. It is not mine. I did not touch it.

## Sample project

- `.tools/samples/dockerplay/docker-compose.yml`, project name `tuiresearch`, 3 services from `nginx:latest`:
  `web` (port 18080), `ticker` (prints a line every second) and `failing` (prints an error and exits with code 3).
- Start: `docker compose up -d`. Stop: `docker compose down` in that folder. The containers are still up at the time of writing.
- lazydocker started with that folder as the working directory, so the Project and Services panels appear.

## Limits of the run

- lazydocker answered very slowly through the driver. A menu opened 7 to 9 seconds after the key.
  The machine had 48 per cent CPU load from other work. I cannot say how much of the delay is lazydocker and how much is load.
- Because of the delay, many captures show the state before my last key. I deleted the captures I could not match to a key.
  The five that remain are named for what they show.
- Mouse clicks in lazydocker were sent, but I could not confirm a result. Mouse support is not verified for this tool.

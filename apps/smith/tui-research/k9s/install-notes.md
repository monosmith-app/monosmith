# k9s install notes

Version run: k9s v0.51.0 (Windows amd64). Platform: Windows 11, native. WSL not used.

## Install

1. Download the release zip: `gh release download v0.51.0 -R derailed/k9s -p k9s_Windows_amd64.zip` into `.tools/dl-k9s/`.
2. Unzip. The binary is `.tools/dl-k9s/k9s.exe`. No installer ran. Nothing is on PATH.

## A cluster to look at

Docker Desktop was installed but stopped. The agent started it (Docker Desktop 28.0.1).
`kind` (latest release, downloaded with gh) was put to `.tools/bin/kind.exe`.

1. First try: `kind create cluster` with the default node image (Kubernetes v1.37.0). It failed at `kubeadm init`. The WSL2 engine runs cgroup v1, and the log warned about it. Fact seen: the cluster did not start. Cause not proven.
2. Second try: `kind create cluster --name tuires --image kindest/node:v1.31.2`. It worked.
3. Demo workload in namespace `demo`: `web` (3 replicas, nginx:alpine), `broken` (bad image tag, ImagePullBackOff), `crash` (exits 1, CrashLoopBackOff), `logger` (prints a line each second), a ConfigMap, a Secret and a Service.
4. After the run the cluster was deleted with `kind delete cluster --name tuires`. Docker Desktop was left running.

## Isolation

The run used these variables so the user profile stayed untouched:

- `KUBECONFIG=.tools/kube/config`
- `K9S_CONFIG_DIR=.tools/kube/k9s`
- `K9S_LOGS_DIR=.tools/kube/k9s-logs`

## Driver

`.tools/drive.py` (shared, created by a sibling agent: ConPTY plus pyte, colour PNG per shot). Command used: `k9s.exe -n demo`, 130 columns by 38 rows.
Screens are real colour PNG renders of the pyte screen, plus `.txt` dumps. They are not terminal-emulator screenshots. Emoji and Nerd Font glyphs may show as boxes.

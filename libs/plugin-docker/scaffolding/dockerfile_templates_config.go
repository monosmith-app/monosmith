package scaffolding

// dockerignore keeps what no image needs out of the build context (the
// workspace): history, installed dependencies and smith's own state. dist/
// stays: an image copies the app's build from it.
const dockerignore = `.git
**/node_modules
.smith
`

// goDockerfile builds the binary inside Docker, so the image is Linux
// whatever the developer's OS, from the whole workspace (go.work and the
// modules the app uses), and ships it alone on distroless.
const goDockerfile = `# The app's image (smith run image). The context is the workspace.
FROM golang:[[.GoVersion]]-alpine AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /out/app ./[[.Dir]]

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/app /app
ENTRYPOINT ["/app"]
`

// nodeDockerfile installs a Node app from its pruned build (smith run prune,
// which the image task runs first): dist/<app dir> holds the build, a
// production package.json and lockfile, and the workspace packages the app
// uses in workspace_modules/. Dependencies first, so a code change reuses
// the install layer.
const nodeDockerfile = `# The app's image (smith run image, after its build and prune). The context
# is the workspace.
FROM node:[[.NodeVersion]]-slim
ENV NODE_ENV=production
WORKDIR /app
COPY dist/[[.Dir]]/package.json dist/[[.Dir]]/package-lock.json ./
COPY dist/[[.Dir]]/workspace_modules ./workspace_modules
RUN npm ci --omit=dev --ignore-scripts && npm cache clean --force
COPY dist/[[.Dir]]/ ./
USER node
CMD ["node", "."]
`

// webDockerfile serves the static build smith writes to dist/<app dir>.
const webDockerfile = `# The app's image (smith run image, after its build). The context is the
# workspace.
FROM nginx:1.29-alpine
COPY dist/[[.Dir]]/ /usr/share/nginx/html/
EXPOSE 80
`

package vite

var reactDevDependencies = map[string]string{
	"@types/react":         "^19.3.0",
	"@types/react-dom":     "^19.3.0",
	"@vitejs/plugin-react": "^6.1.2",
}

// reactOverrides settle a peer conflict that makes npm refuse to install.
// @vitejs/plugin-react optionally peers @rolldown/plugin-babel, which
// optionally peers @babel/plugin-transform-runtime ^7 || ^8; the newest is 8,
// which wants @babel/core 8. When the workspace already holds Babel 7 (Jest
// brings it) npm cannot place both and fails with ERESOLVE (issue 584). The
// Babel the Jest stack brings is 7, so the workspace stays on 7.
var reactOverrides = map[string]string{"@babel/plugin-transform-runtime": "^7.29.0"}

// reactMain mounts the app slice's root component.
const reactMain = `import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { App } from './app/index.ts'

const root = document.getElementById('app')
if (root === null) {
  throw new Error('index.html has no #app')
}
createRoot(root).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
`

// react is a React app: the app slice's root component greets.
var react = webApp{
	Entry: "main.tsx", Typecheck: "tsc -p tsconfig.json --noEmit", JSX: true,
	PluginImport: "import react from '@vitejs/plugin-react'", PluginCall: "react()",
	Dependencies:    map[string]string{"react": "^19.3.1", "react-dom": "^19.3.1"},
	DevDependencies: reactDevDependencies,
	Overrides:       reactOverrides,
	Files: map[string]string{
		"[[.Dir]]/src/main.tsx": reactMain,
		"[[.Dir]]/src/app/app.component.tsx": `import { greet } from '../greeting'

/** The app's root component. */
export function App () {
  return <h1>{greet('world').message}</h1>
}
`,
		"[[.Dir]]/src/app/index.ts": `export { App } from './app.component.tsx'
`,
	},
}

// reactP5 is a React app that draws with p5: a bouncing-ball slice whose
// motion is a pure algorithm, whose sketch is the p5 client, and whose
// component loads p5 lazily.
var reactP5 = webApp{
	Entry: "main.tsx", Typecheck: "tsc -p tsconfig.json --noEmit", JSX: true,
	PluginImport: "import react from '@vitejs/plugin-react'", PluginCall: "react()",
	Dependencies:    map[string]string{"@p5-wrapper/react": "^5.0.4", "p5": "^2.3.4", "react": "^19.3.1", "react-dom": "^19.3.1"},
	DevDependencies: reactDevDependencies,
	Overrides:       reactOverrides,
	Files: map[string]string{
		"[[.Dir]]/src/main.tsx": reactMain,
		"[[.Dir]]/src/app/app.component.tsx": `import { BouncingBall } from '../bouncing-ball'
import { greet } from '../greeting'

/** The app's root component. */
export function App () {
  return (
    <main>
      <h1>{greet('world').message}</h1>
      <BouncingBall />
    </main>
  )
}
`,
		"[[.Dir]]/src/app/index.ts": `export { App } from './app.component.tsx'
`,
		"[[.Dir]]/src/bouncing-ball/bounce.algorithm.ts": `/** A ball's position and velocity along one axis. */
export interface Motion {
  position: number
  velocity: number
}

/**
 * Moves a ball one step, bouncing it off the walls at 0 and size.
 * @param motion - where the ball is and how fast it moves
 * @param size - the distance between the walls
 * @returns the ball's next motion
 */
export function bounce (motion: Motion, size: number): Motion {
  const position = motion.position + motion.velocity
  if (position < 0 || position > size) {
    return { position: motion.position - motion.velocity, velocity: -motion.velocity }
  }

  return { position, velocity: motion.velocity }
}
`,
		"[[.Dir]]/src/bouncing-ball/bouncing-ball.client.ts": `import type { Sketch } from '@p5-wrapper/react'
import { bounce, type Motion } from './bounce.algorithm.ts'

const size = 400

/** The p5 sketch: a ball bouncing around the canvas. */
export const bouncingBall: Sketch = (p5) => {
  let x: Motion = { position: size / 2, velocity: 3 }
  let y: Motion = { position: size / 3, velocity: 2 }
  p5.setup = () => {
    p5.createCanvas(size, size)
  }
  p5.draw = () => {
    x = bounce(x, size)
    y = bounce(y, size)
    p5.background(240)
    p5.circle(x.position, y.position, 24)
  }
}
`,
		"[[.Dir]]/src/bouncing-ball/bouncing-ball.component.tsx": `import { P5Canvas } from '@p5-wrapper/react'
import { Suspense } from 'react'
import { bouncingBall } from './bouncing-ball.client.ts'

/** The bouncing ball, drawn by p5 once it has loaded. */
export function BouncingBall () {
  return (
    <Suspense fallback={<p>Loading…</p>}>
      <P5Canvas sketch={bouncingBall} />
    </Suspense>
  )
}
`,
		"[[.Dir]]/src/bouncing-ball/index.ts": `export { BouncingBall } from './bouncing-ball.component.tsx'
`,
	},
}

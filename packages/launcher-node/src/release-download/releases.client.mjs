/** Where smith's release archives are by default. */
const GITHUB = 'https://github.com/monosmith-app/monosmith/releases/download'
const API = 'https://api.github.com/repos/monosmith-app/monosmith/releases?per_page=100'

/**
 * The address a release's files are under: SMITH_DOWNLOAD_URL (a mirror), else GitHub.
 *
 * @param {NodeJS.ProcessEnv} env - The environment.
 * @returns {string} The base, without a trailing slash.
 */
export function downloadBase (env) {
  return (env.SMITH_DOWNLOAD_URL || GITHUB).replace(/\/+$/, '')
}

/**
 * Fetches a URL and fails with a sentence that names it.
 *
 * @param {string} url - The address.
 * @param {RequestInit} [init] - Request options.
 * @returns {Promise<Response>} The answer.
 * @throws {Error} When the server cannot be reached or does not answer 200.
 */
async function get (url, init) {
  let res
  try {
    res = await fetch(url, init)
  } catch (error) {
    throw new Error(`cannot reach ${new URL(url).host} (${error.cause?.code ?? error.message}); behind a firewall, set SMITH_DOWNLOAD_URL to a mirror of smith's release archives`, { cause: error })
  }
  if (!res.ok) throw new Error(`${url} answered ${res.status}`)

  return res
}

/**
 * One file of a release: <base>/smith@<version>/<name>.
 *
 * @param {NodeJS.ProcessEnv} env - The environment.
 * @param {string} version - An exact version.
 * @param {string} name - The file's name.
 * @returns {Promise<Buffer>} Its bytes.
 */
export async function releaseFile (env, version, name) {
  const res = await get(`${downloadBase(env)}/smith%40${version}/${name}`)

  return Buffer.from(await res.arrayBuffer())
}

/**
 * The newest stable smith: the mirror's latest.txt, or the newest smith@ release GitHub lists
 * (GITHUB_TOKEN or GH_TOKEN, when set, raises GitHub's rate limit).
 *
 * @param {NodeJS.ProcessEnv} env - The environment.
 * @returns {Promise<string>} The version.
 * @throws {Error} When no release can be found.
 */
export async function newestRelease (env) {
  if (env.SMITH_DOWNLOAD_URL) {
    const res = await get(`${downloadBase(env)}/latest.txt`)
    const text = await res.text()

    return text.trim()
  }
  const token = env.GITHUB_TOKEN || env.GH_TOKEN
  const res = await get(API, { headers: { Accept: 'application/vnd.github+json', ...(token && { Authorization: `Bearer ${token}` }) } })
  const releases = await res.json()
  const newest = releases.find(r => r.tag_name?.startsWith('smith@') && !r.prerelease && !r.draft)
  if (!newest) throw new Error('GitHub lists no smith release')

  return newest.tag_name.slice('smith@'.length)
}

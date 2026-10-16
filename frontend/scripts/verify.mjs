/* ============================================================================
   Verifikasi headless FE (CDP) — login via API (cookie), navigasi, screenshot,
   kumpulkan error konsol. Dipakai untuk memverifikasi integrasi tiap task.

   Pakai:
     node scripts/verify.mjs <path> <nama> [--as=bk|wk|anon|none] [--eval="<js>"] [--w=1440]

   Prasyarat: backend jalan di :8080, dev server FE jalan di :5173 (pnpm dev).
   ========================================================================== */
import { spawn } from 'node:child_process'
import { existsSync, mkdirSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

const CANDIDATES = [
  '/Users/whm/.cache/puppeteer/chrome-headless-shell/mac_arm-150.0.7871.24/chrome-headless-shell-mac-arm64/chrome-headless-shell',
  '/Users/whm/.cache/puppeteer/chrome-headless-shell/mac_arm-148.0.7778.97/chrome-headless-shell-mac-arm64/chrome-headless-shell',
  '/Users/whm/.cache/puppeteer/chrome-headless-shell/mac_arm-134.0.6998.35/chrome-headless-shell-mac-arm64/chrome-headless-shell',
]
const CHROME = CANDIDATES.find(existsSync)
if (!CHROME) {
  console.error('chrome-headless-shell tidak ditemukan')
  process.exit(2)
}

const path = process.argv[2] || '/'
const name = process.argv[3] || 'page'
const asArg = (process.argv.find((a) => a.startsWith('--as=')) || '').split('=')[1] || 'none'
const evalArg = process.argv.find((a) => a.startsWith('--eval='))
const evalJs = evalArg ? evalArg.slice('--eval='.length) : null
const width = Number((process.argv.find((a) => a.startsWith('--w=')) || '').split('=')[1]) || 1440
const height = Number((process.argv.find((a) => a.startsWith('--h=')) || '').split('=')[1]) || 1024
// 127.0.0.1 (bukan localhost): Node bisa resolve localhost → ::1 (IPv6),
// sedangkan server Go listen di IPv4 saja.
const origin = process.env.ORIGIN || 'http://localhost:5173'
const api = process.env.API || 'http://127.0.0.1:8080'
const url = origin + path

const outDir = join(process.cwd(), '.verify')
if (!existsSync(outDir)) mkdirSync(outDir, { recursive: true })
const shot = join(outDir, `${name}.png`)
const profile = join(tmpdir(), 'fe-verify-' + process.pid)

const child = spawn(
  CHROME,
  [
    '--headless',
    '--no-sandbox',
    '--disable-gpu',
    '--hide-scrollbars',
    `--window-size=${width},${height}`,
    '--remote-debugging-port=0',
    `--user-data-dir=${profile}`,
    'about:blank',
  ],
  { stdio: ['ignore', 'pipe', 'pipe'] },
)

let browserWs = null
const waitWs = new Promise((resolve) => {
  child.stderr.on('data', (d) => {
    const m = d.toString().match(/DevTools listening on (ws:\/\/\S+)/)
    if (m && !browserWs) {
      browserWs = m[1]
      resolve(browserWs)
    }
  })
  setTimeout(() => resolve(null), 10000)
})

async function resolvePageTarget(bws) {
  const m = bws.match(/ws:\/\/([^/]+)\//)
  const host = m ? m[1] : null
  if (!host) return null
  for (let i = 0; i < 20; i++) {
    try {
      const res = await fetch('http://' + host + '/json/list')
      const list = await res.json()
      const page = list.find((t) => t.type === 'page' && t.webSocketDebuggerUrl)
      if (page) return page.webSocketDebuggerUrl
    } catch {
      /* retry */
    }
    await new Promise((r) => setTimeout(r, 300))
  }
  return null
}

const errors = []
let msgId = 0
const pending = new Map()

function send(ws, method, params = {}) {
  const id = ++msgId
  ws.send(JSON.stringify({ id, method, params }))
  return new Promise((resolve) => pending.set(id, resolve))
}
const sleep = (ms) => new Promise((r) => setTimeout(r, ms))

/** Login via API → kembalikan header Cookie untuk dipasang di browser. */
async function loginCookie(email) {
  const res = await fetch(`${api}/api/v1/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password: 'password123' }),
  })
  if (!res.ok) return null
  const setCookie = res.headers.getSetCookie?.() ?? []
  return setCookie.map((c) => c.split(';')[0]).join('; ')
}

async function main() {
  const bws = await waitWs
  if (!bws) {
    console.error('gagal konek CDP')
    child.kill('SIGKILL')
    process.exit(2)
  }
  const ws = await resolvePageTarget(bws)
  const sock = new WebSocket(ws)
  await new Promise((r) => {
    sock.onopen = r
  })

  sock.onmessage = (ev) => {
    const m = JSON.parse(ev.data)
    if (m.id && pending.has(m.id)) {
      pending.get(m.id)(m.result)
      pending.delete(m.id)
      return
    }
    if (m.method === 'Runtime.consoleAPICalled' && (m.params.type === 'error' || m.params.type === 'assert')) {
      errors.push(m.params.args.map((a) => a.value || a.description || '').join(' '))
    }
    if (m.method === 'Runtime.exceptionThrown') {
      const d = m.params.exceptionDetails
      errors.push((d.exception && (d.exception.description || d.exception.value)) || d.text)
    }
    if (m.method === 'Log.entryAdded' && m.params.entry.level === 'error') {
      errors.push(m.params.entry.text)
    }
  }

  await send(sock, 'Runtime.enable')
  await send(sock, 'Log.enable')
  await send(sock, 'Page.enable')
  await send(sock, 'Network.enable')

  // Navigasi dulu ke origin (butuh halaman agar cookie terpasang di origin benar).
  await send(sock, 'Page.navigate', { url: origin + '/login' })
  await sleep(800)

  // Pasang cookie sesi via API (tanpa lewat UI login).
  // Cookie anon siswa bisa diberikan via env ANON_COOKIE (mis. "anon_token=...").
  let cookie = null
  if (asArg === 'bk') cookie = await loginCookie('bk@demo.local')
  else if (asArg === 'wk') cookie = await loginCookie('wali@demo.local')
  else if (process.env.ANON_COOKIE) cookie = process.env.ANON_COOKIE
  if (cookie) {
    // Cookie bisa "name=value" tunggal (auth_token atau anon_token).
    const eq = cookie.indexOf('=')
    const cname = eq > 0 ? cookie.slice(0, eq) : 'auth_token'
    const cval = eq > 0 ? cookie.slice(eq + 1) : cookie
    const setRes = await send(sock, 'Network.setCookie', {
      name: cname,
      value: cval,
      url: origin,
      path: '/',
    })
    if (setRes && setRes.success === false) console.error('setCookie gagal:', JSON.stringify(setRes))
  }

  await send(sock, 'Page.navigate', { url })
  await sleep(1800)

  if (evalJs) {
    const r = await send(sock, 'Runtime.evaluate', { expression: evalJs, awaitPromise: true, returnByValue: true })
    if (r && r.exceptionDetails) errors.push('EVAL: ' + (r.exceptionDetails.text || ''))
    if (r && r.result && r.result.value !== undefined) console.log('EVAL →', JSON.stringify(r.result.value))
    await sleep(900)
  }

  const shotRes = await send(sock, 'Page.captureScreenshot', { format: 'png' })
  if (shotRes && shotRes.data) {
    const { writeFileSync } = await import('node:fs')
    writeFileSync(shot, Buffer.from(shotRes.data, 'base64'))
    console.log('screenshot:', shot)
  }

  // Cek overflow horizontal.
  const ov = await send(sock, 'Runtime.evaluate', {
    expression: 'document.documentElement.scrollWidth - document.documentElement.clientWidth',
    returnByValue: true,
  })
  console.log('h-overflow:', ov?.result?.value ?? '?')

  console.log('errors:', errors.length ? errors : 'none')
  child.kill('SIGKILL')
  process.exit(errors.length ? 1 : 0)
}

main().catch((e) => {
  console.error(e)
  child.kill('SIGKILL')
  process.exit(2)
})

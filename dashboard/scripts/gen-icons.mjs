// Generates build/icon.ico (multi-size) for NetController.
// Command-center logo: dark rounded square + glowing cyan diamond mark.
// Pure-JS raster (pngjs) with 4x supersampling, then packed to .ico via png-to-ico.
import { PNG } from 'pngjs'
import pngToIco from 'png-to-ico'
import { writeFileSync, mkdirSync } from 'fs'
import { fileURLToPath } from 'url'
import { dirname, join } from 'path'

const root = join(dirname(fileURLToPath(import.meta.url)), '..')
const SS = 4 // supersample factor
const SIZES = [256, 128, 64, 48, 32, 24, 16]

function mix(a, b, t) { return a + (b - a) * t }
function clamp01(x) { return x < 0 ? 0 : x > 1 ? 1 : x }
function smooth(edge0, edge1, x) {
  const t = clamp01((x - edge0) / (edge1 - edge0))
  return t * t * (3 - 2 * t)
}

// Signed distance to a rounded box centred in unit square. p,half,r in [0..1].
function sdRoundBox(px, py, hx, hy, r) {
  const qx = Math.abs(px) - hx + r
  const qy = Math.abs(py) - hy + r
  const ax = Math.max(qx, 0), ay = Math.max(qy, 0)
  return Math.sqrt(ax * ax + ay * ay) + Math.min(Math.max(qx, qy), 0) - r
}

function renderSize(size) {
  const S = size * SS
  const png = new PNG({ width: S, height: S })
  for (let y = 0; y < S; y++) {
    for (let x = 0; x < S; x++) {
      // normalised coords in [-0.5, 0.5]
      const u = (x + 0.5) / S - 0.5
      const v = (y + 0.5) / S - 0.5
      const px = S > 0 ? 1 / S : 0 // pixel size in uv

      // --- background rounded square ---
      const dBox = sdRoundBox(u, v, 0.5, 0.5, 0.22)
      const bgCov = smooth(px, -px, dBox)
      // vertical gradient dark
      const gt = clamp01((v + 0.5))
      let r = mix(0x0c, 0x05, gt)
      let g = mix(0x13, 0x08, gt)
      let b = mix(0x1d, 0x0d, gt)
      // subtle inner border ring
      const ring = smooth(0.030, 0.0, Math.abs(dBox + 0.02)) * 0.5
      r = mix(r, 0x78, ring * 0.5); g = mix(g, 0xa8, ring * 0.5); b = mix(b, 0xd2, ring * 0.5)

      // --- glowing cyan diamond (L1 ball) ---
      const dia = (Math.abs(u) + Math.abs(v)) - 0.30 // <0 inside
      // glow outside the diamond
      const glow = Math.exp(-Math.max(dia, 0) * 11) * 0.85
      r = mix(r, 0x36, glow * 0.55 * bgCov)
      g = mix(g, 0xd3, glow * 0.85 * bgCov)
      b = mix(b, 0xee, glow * 0.95 * bgCov)
      // diamond fill (gradient top->bottom) with AA edge
      const diaCov = smooth(px * 1.5, -px * 1.5, dia)
      const dt = clamp01((v + 0.30) / 0.60)
      const fr = mix(0x6c, 0x22, dt)
      const fg = mix(0xeb, 0xb3, dt)
      const fb = mix(0xf7, 0xd6, dt)
      r = mix(r, fr, diaCov); g = mix(g, fg, diaCov); b = mix(b, fb, diaCov)
      // top-left highlight sheen on the diamond
      const sheen = smooth(0.0, -0.10, dia) * smooth(0.10, -0.25, u + v)
      r = mix(r, 0xff, sheen * 0.35 * diaCov)
      g = mix(g, 0xff, sheen * 0.35 * diaCov)
      b = mix(b, 0xff, sheen * 0.35 * diaCov)

      const idx = (S * y + x) << 2
      png.data[idx] = Math.round(clamp01(r / 255) * 255)
      png.data[idx + 1] = Math.round(clamp01(g / 255) * 255)
      png.data[idx + 2] = Math.round(clamp01(b / 255) * 255)
      png.data[idx + 3] = Math.round(bgCov * 255)
    }
  }
  // downsample SS -> 1 (box filter)
  const out = new PNG({ width: size, height: size })
  for (let y = 0; y < size; y++) {
    for (let x = 0; x < size; x++) {
      let r = 0, g = 0, b = 0, a = 0
      for (let sy = 0; sy < SS; sy++) {
        for (let sx = 0; sx < SS; sx++) {
          const i = (S * (y * SS + sy) + (x * SS + sx)) << 2
          const af = png.data[i + 3]
          r += png.data[i] * af; g += png.data[i + 1] * af; b += png.data[i + 2] * af; a += af
        }
      }
      const oi = (size * y + x) << 2
      out.data[oi] = a ? Math.round(r / a) : 0
      out.data[oi + 1] = a ? Math.round(g / a) : 0
      out.data[oi + 2] = a ? Math.round(b / a) : 0
      out.data[oi + 3] = Math.round(a / (SS * SS))
    }
  }
  return PNG.sync.write(out)
}

mkdirSync(join(root, 'build'), { recursive: true })
const buffers = SIZES.map(renderSize)
// keep a 256 png too (handy for docs / linux)
writeFileSync(join(root, 'build', 'icon.png'), buffers[0])
const ico = await pngToIco(buffers)
writeFileSync(join(root, 'build', 'icon.ico'), ico)
console.log('✓ build/icon.ico (' + SIZES.join(',') + ') + build/icon.png')

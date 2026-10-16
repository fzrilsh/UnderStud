import { useEffect, useRef } from 'react'

import type { ChartPoint } from '@/features/dashboard/hooks'

const SERIES = [
  { key: 'pinned', name: 'Disematkan', color: '#ef4444' },
  { key: 'drafted', name: 'Draft', color: '#16a34a' },
  { key: 'queued', name: 'Antrean', color: '#3d3ba6' },
] as const

/**
 * Line chart sederhana (port mockup/js/chart.js) — 3 seri, grid + label sumbu.
 * Menggambar ulang saat ukuran elemen berubah (ResizeObserver).
 */
export default function MiniChart({ points, showLegend = false }: { points: ChartPoint[]; showLegend?: boolean }) {
  const canvasRef = useRef<HTMLCanvasElement>(null)

  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) return

    const draw = () => {
      const dpr = window.devicePixelRatio || 1
      const cssW = canvas.clientWidth || 520
      const cssH = canvas.clientHeight || 220
      if (!cssW || !cssH) return
      canvas.width = Math.round(cssW * dpr)
      canvas.height = Math.round(cssH * dpr)
      const ctx = canvas.getContext('2d')
      if (!ctx) return
      ctx.setTransform(dpr, 0, 0, dpr, 0, 0)
      ctx.clearRect(0, 0, cssW, cssH)

      const labels = points.map((p) => p.label)
      const series = SERIES.map((s) => ({
        name: s.name,
        color: s.color,
        data: points.map((p) => Number(p[s.key] ?? 0)),
      }))

      const padL = 34
      const padR = 12
      const padT = 12
      const padB = 26
      const plotW = cssW - padL - padR
      const plotH = cssH - padT - padB

      let max = 1
      series.forEach((s) => s.data.forEach((v) => { if (v > max) max = v }))
      max = Math.ceil(max / 20) * 20 || 20

      const xAt = (i: number) => padL + (labels.length <= 1 ? plotW / 2 : (plotW * i) / (labels.length - 1))
      const yAt = (v: number) => padT + plotH - (v / max) * plotH

      // Grid + label Y
      ctx.font = '11px "Plus Jakarta Sans", sans-serif'
      ctx.fillStyle = 'rgba(21,21,58,.4)'
      const steps = 4
      for (let i = 0; i <= steps; i++) {
        const v = (max / steps) * i
        const y = yAt(v)
        ctx.strokeStyle = 'rgba(21,21,58,.08)'
        ctx.beginPath()
        ctx.moveTo(padL, y)
        ctx.lineTo(cssW - padR, y)
        ctx.stroke()
        ctx.fillText(String(Math.round(v)), 4, y + 4)
      }

      // Label X
      ctx.fillStyle = 'rgba(21,21,58,.4)'
      labels.forEach((lb, i) => {
        const x = xAt(i)
        const w = ctx.measureText(lb).width
        ctx.fillText(lb, x - w / 2, cssH - 8)
      })

      // Area fill (seri pertama saja, seperti mockup) + garis + titik putih.
      const hexA = (hex: string, a: number) => {
        const n = parseInt(hex.slice(1), 16)
        return `rgba(${(n >> 16) & 255},${(n >> 8) & 255},${n & 255},${a})`
      }
      series.forEach((s, si) => {
        if (si === 0) {
          ctx.beginPath()
          s.data.forEach((v, i) => {
            const x = xAt(i)
            const y = yAt(v)
            if (i === 0) ctx.moveTo(x, y)
            else ctx.lineTo(x, y)
          })
          ctx.lineTo(xAt(s.data.length - 1), padT + plotH)
          ctx.lineTo(xAt(0), padT + plotH)
          ctx.closePath()
          const g = ctx.createLinearGradient(0, padT, 0, padT + plotH)
          g.addColorStop(0, hexA(s.color, 0.18))
          g.addColorStop(1, hexA(s.color, 0.01))
          ctx.fillStyle = g
          ctx.fill()
        }
        // Garis
        ctx.beginPath()
        ctx.lineWidth = 2
        ctx.strokeStyle = s.color
        s.data.forEach((v, i) => {
          const x = xAt(i)
          const y = yAt(v)
          if (i === 0) ctx.moveTo(x, y)
          else ctx.lineTo(x, y)
        })
        ctx.stroke()
        // Titik: isi putih + border warna seri (mockup)
        s.data.forEach((v, i) => {
          ctx.beginPath()
          ctx.arc(xAt(i), yAt(v), 3, 0, Math.PI * 2)
          ctx.fillStyle = '#fff'
          ctx.fill()
          ctx.lineWidth = 2
          ctx.strokeStyle = s.color
          ctx.stroke()
        })
      })
    }

    draw()
    const raf = requestAnimationFrame(draw)
    const t1 = setTimeout(draw, 120)
    const t2 = setTimeout(draw, 400)
    let ro: ResizeObserver | undefined
    if (window.ResizeObserver) {
      ro = new ResizeObserver(draw)
      ro.observe(canvas)
    }
    return () => {
      cancelAnimationFrame(raf)
      clearTimeout(t1)
      clearTimeout(t2)
      ro?.disconnect()
    }
  }, [points])

  return (
    <div className="flex h-full min-h-0 flex-col gap-3">
      {showLegend && (
        <div className="flex flex-wrap items-center gap-4">
          {SERIES.map((s) => (
            <span key={s.key} className="flex items-center gap-2 text-[12px] text-[rgba(21,21,58,.8)]">
              <span className="size-2.5 rounded-full" style={{ background: s.color }} />
              {s.name}
            </span>
          ))}
        </div>
      )}
      <canvas ref={canvasRef} className="min-h-0 w-full flex-1" />
    </div>
  )
}

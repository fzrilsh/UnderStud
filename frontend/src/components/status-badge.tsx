import { cn } from '@/lib/utils'

type Tone = 'green' | 'red' | 'redDark' | 'orange' | 'violet' | 'gray'

const TONE: Record<Tone, { bg: string; dot: string; text: string }> = {
  green: { bg: 'bg-[#dcfce7]', dot: 'bg-[#16a34a]', text: 'text-[#15153a]' },
  red: { bg: 'bg-[#fee2e2]', dot: 'bg-[#dc2626]', text: 'text-[#ef4444]' },
  /** Sama seperti `red` tapi teks gelap (dipakai badge jumlah red flag di dashboard). */
  redDark: { bg: 'bg-[#fee2e2]', dot: 'bg-[#dc2626]', text: 'text-[#15153a]' },
  orange: { bg: 'bg-[#fef9c3]', dot: 'bg-[#ea580c]', text: 'text-[#ea580c]' },
  violet: { bg: 'bg-[#ecebf6]', dot: 'bg-[#3d3ba6]', text: 'text-[#3d3ba6]' },
  gray: { bg: 'bg-[#f1f1f5]', dot: 'bg-[#a1a1aa]', text: 'text-[rgba(21,21,58,.6)]' },
}

/** Badge status dengan dot (padanan UI.badge mockup). */
export default function StatusBadge({ tone, children }: { tone: Tone; children: React.ReactNode }) {
  const t = TONE[tone]
  return (
    <span className={cn('inline-flex items-center gap-2 rounded-[6px] px-2 py-[3px]', t.bg)}>
      <span className={cn('size-2.5 rounded-full', t.dot)} />
      <span className={cn('text-[12px] font-semibold leading-[1.5]', t.text)}>{children}</span>
    </span>
  )
}

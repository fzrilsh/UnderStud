import { createContext, useCallback, useContext, useRef, useState, type ReactNode } from 'react'

import { Button } from '@/components/ui/button'
import { IconClose } from '@/components/icons'
import { cn } from '@/lib/utils'

interface ConfirmOptions {
  title: string
  body?: string
  confirmText?: string
  cancelText?: string
  /** Tombol konfirmasi merah (aksi berbahaya). */
  danger?: boolean
}

type ConfirmFn = (opts: ConfirmOptions) => Promise<boolean>

const ConfirmContext = createContext<ConfirmFn | null>(null)

/** Dialog konfirmasi bergaya mockup (Promise<boolean>). */
export function ConfirmProvider({ children }: { children: ReactNode }) {
  const [opts, setOpts] = useState<ConfirmOptions | null>(null)
  const resolver = useRef<((v: boolean) => void) | null>(null)

  const confirm = useCallback<ConfirmFn>((o) => {
    setOpts(o)
    return new Promise<boolean>((resolve) => {
      resolver.current = resolve
    })
  }, [])

  const close = (value: boolean) => {
    resolver.current?.(value)
    resolver.current = null
    setOpts(null)
  }

  return (
    <ConfirmContext.Provider value={confirm}>
      {children}
      {opts && (
        <div className="fixed inset-0 z-[100] flex items-center justify-center bg-[rgba(21,21,58,.45)] p-4">
          <div className="relative w-full max-w-[440px] rounded-2xl bg-white p-6 shadow-xl">
            <button
              type="button"
              onClick={() => close(false)}
              className="absolute right-4 top-4 flex size-7 items-center justify-center rounded-md text-violet-dark hover:bg-violet-light/60"
              title="Tutup"
            >
              <IconClose />
            </button>
            <h3 className="font-heading text-[20px] font-semibold text-violet-dark">{opts.title}</h3>
            {opts.body && <p className="mt-2 text-[14px] text-[rgba(21,21,58,.6)]">{opts.body}</p>}
            <div className="mt-6 flex justify-end gap-3">
              <Button
                variant="outline"
                onClick={() => close(false)}
                className="border-violet-light text-violet-dark"
              >
                {opts.cancelText ?? 'Batal'}
              </Button>
              <Button
                onClick={() => close(true)}
                className={cn(opts.danger ? 'bg-[#B42318] hover:bg-[#8f1c13]' : 'bg-violet-dark hover:bg-violet-hover')}
              >
                {opts.confirmText ?? 'Konfirmasi'}
              </Button>
            </div>
          </div>
        </div>
      )}
    </ConfirmContext.Provider>
  )
}

export function useConfirm(): ConfirmFn {
  const ctx = useContext(ConfirmContext)
  if (!ctx) throw new Error('useConfirm harus dipakai di dalam <ConfirmProvider>')
  return ctx
}

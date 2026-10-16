import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Link } from 'react-router-dom'
import { Eye, Loader2 } from 'lucide-react'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useLogin } from '@/features/auth/hooks'
import { loginSchema, type LoginInput } from '@/features/auth/schemas'

export default function LoginPage() {
  const login = useLogin()
  const [showPw, setShowPw] = useState(false)
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginInput>({
    resolver: zodResolver(loginSchema),
    defaultValues: { email: '', password: '' },
  })

  return (
    <main className="flex min-h-svh items-center justify-center bg-base px-4 py-10">
      <div className="w-full max-w-[420px] rounded-2xl border border-violet-light bg-white p-8 shadow-[0_24px_60px_rgba(21,21,58,.10)] md:p-9">
        <div className="flex flex-col items-center">
          <img src="/understud-logo.png" alt="UnderStud" className="size-10 object-contain" />
          <h1 className="text-gradient-heading mt-4 font-heading text-[34px] font-semibold leading-tight">
            Selamat datang
          </h1>
          <p className="mt-1.5 text-[13.5px] text-[rgba(21,21,58,.6)]">
            Masuk dengan akun staf sekolah
          </p>
        </div>

        <form onSubmit={handleSubmit((data) => login.mutate(data))} className="mt-8 flex flex-col gap-4">
          <div>
            <Label htmlFor="email" className="mb-1.5 text-[13px] font-semibold">
              Email
            </Label>
            <Input
              id="email"
              type="email"
              autoComplete="username"
              placeholder="Masukkan email"
              {...register('email')}
            />
            {errors.email && <p className="mt-1 text-[12.5px] text-destructive">{errors.email.message}</p>}
          </div>

          <div>
            <Label htmlFor="password" className="mb-1.5 text-[13px] font-semibold">
              Kata sandi
            </Label>
            <div className="relative">
              <Input
                id="password"
                type={showPw ? 'text' : 'password'}
                autoComplete="current-password"
                placeholder="Masukkan kata sandi"
                className="pr-11"
                {...register('password')}
              />
              <button
                type="button"
                title="Tampilkan kata sandi"
                onClick={() => setShowPw((v) => !v)}
                className="absolute right-2 top-1/2 flex size-8 -translate-y-1/2 items-center justify-center rounded-md hover:bg-violet-light/60"
              >
                <Eye className="size-[18px] text-violet-dark" />
              </button>
            </div>
            {errors.password && (
              <p className="mt-1 text-[12.5px] text-destructive">{errors.password.message}</p>
            )}
          </div>

          <Button type="submit" disabled={login.isPending} className="mt-1 w-full bg-violet-dark hover:bg-violet-hover">
            {login.isPending && <Loader2 className="size-4 animate-spin" />}
            Masuk
          </Button>
        </form>

        <p className="mt-4 text-center text-[12.5px] text-[rgba(21,21,58,.55)]">
          Tidak bisa masuk? Hubungi admin sekolah.
        </p>

        <div className="mt-6 border-t border-violet-light pt-5 text-center text-[13px] text-[rgba(21,21,58,.7)]">
          Kamu siswa?{' '}
          <Link to="/lapor" className="font-semibold text-violet underline underline-offset-2 hover:text-violet-hover">
            Kirim laporan anonim
          </Link>
        </div>
      </div>
    </main>
  )
}

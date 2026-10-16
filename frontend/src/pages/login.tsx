import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Link } from 'react-router-dom'

import { IconEye } from '@/components/icons'
import { Button } from '@/components/ui/button'
import { useLogin } from '@/features/auth/hooks'
import { getApiErrorMessage } from '@/lib/errors'
import { loginSchema, type LoginInput } from '@/features/auth/schemas'
import { useSeo } from '@/lib/seo'

export default function LoginPage() {
  useSeo('Masuk', 'Masuk ke akun staf sekolah untuk meninjau laporan siswa dan data pendampingan.')

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

  // Error server ditampilkan inline di atas tombol (mockup `#login-error`).
  const serverError = login.isError ? getApiErrorMessage(login.error, 'Email atau kata sandi tidak cocok.') : null

  return (
    <main className="flex min-h-svh items-center justify-center bg-base px-4 py-10">
      {/* Dekorasi lingkaran latar (mockup login/index.html). */}
      <div className="pointer-events-none fixed inset-0 -z-10">
        <img
          src="/circle-a.svg"
          alt=""
          className="absolute -left-[10%] -top-[20%] w-[45%] rotate-180 opacity-[0.05]"
        />
        <img src="/circle-b.svg" alt="" className="absolute -bottom-[15%] -right-[8%] w-[42%] opacity-[0.10]" />
      </div>

      <div className="w-full max-w-[420px] rounded-[16px] border border-[rgba(226,226,242,.9)] bg-white p-8 shadow-[0_24px_60px_rgba(21,21,58,.10)] md:p-9">
        <div className="flex flex-col items-center">
          <img src="/understud-logo.png" alt="UnderStud" className="size-10 object-contain" />
          <h1 className="text-gradient-heading mt-4 font-heading text-[34px] font-semibold leading-tight">
            Selamat datang
          </h1>
          <p className="mt-1.5 text-[13.5px] text-[rgba(21,21,58,.6)]">
            Masuk dengan akun staf sekolah
          </p>
        </div>

        <form noValidate onSubmit={handleSubmit((data) => login.mutate(data))} className="mt-8 flex flex-col gap-4">
          <div>
            <label htmlFor="email" className="mb-1.5 block text-[13px] font-semibold text-violet-dark">
              Email
            </label>
            <input
              id="email"
              type="email"
              autoComplete="username"
              placeholder="Masukkan email"
              className="input"
              {...register('email')}
            />
          </div>

          <div>
            <label htmlFor="password" className="mb-1.5 block text-[13px] font-semibold text-violet-dark">
              Kata sandi
            </label>
            <div className="relative">
              <input
                id="password"
                type={showPw ? 'text' : 'password'}
                autoComplete="current-password"
                placeholder="Masukkan kata sandi"
                className="input pr-11"
                {...register('password')}
              />
              <button
                type="button"
                title="Tampilkan kata sandi"
                onClick={() => setShowPw((v) => !v)}
                className="absolute right-2 top-1/2 flex size-8 -translate-y-1/2 items-center justify-center rounded-md hover:bg-violet-light/60"
              >
                <IconEye className="text-violet-dark" />
              </button>
            </div>
          </div>

          {(serverError || errors.email || errors.password) && (
            <p className="text-[13px] font-medium text-[#B42318]">
              {serverError ?? errors.email?.message ?? errors.password?.message}
            </p>
          )}

          <Button
            type="submit"
            disabled={login.isPending}
            className="mt-1 h-10 w-full rounded-[10px] bg-violet px-4 text-[14px] font-semibold text-base hover:bg-violet-hover"
          >
            {login.isPending ? 'Memproses…' : 'Masuk'}
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

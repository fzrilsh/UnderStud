/** Ikon garis mockup (diport dari mockup/js/components.js). Stroke #15153A. */
import type { SVGProps } from 'react'

type P = SVGProps<SVGSVGElement>

export const IconMenu = (p: P) => (
  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" {...p}>
    <path d="M4 7h16M4 12h16M4 17h16" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" />
  </svg>
)

export const IconClose = (p: P) => (
  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" {...p}>
    <path d="M6 6l12 12M18 6L6 18" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" />
  </svg>
)

export const IconChevrons = (p: P) => (
  <svg width="24" height="24" viewBox="0 0 24 24" fill="none" style={{ transform: 'rotate(-90deg)' }} {...p}>
    <path d="M12.636 6.8284L7.6863 11.7782L6.27209 10.364L12.636 4L19 10.364L17.5858 11.7782L12.636 6.8284Z" fill="currentColor" />
    <path d="M12.636 13.8284L7.6863 18.7782L6.27209 17.364L12.636 11L19 17.364L17.5858 18.7782L12.636 13.8284Z" fill="currentColor" />
  </svg>
)

export const IconDashboard = (p: P) => (
  <svg width="18" height="18" viewBox="0 0 18 18" fill="none" {...p}>
    <rect x="2" y="2" width="6" height="6" rx="1.5" stroke="currentColor" strokeWidth="1.5" />
    <rect x="10" y="2" width="6" height="6" rx="1.5" stroke="currentColor" strokeWidth="1.5" />
    <rect x="2" y="10" width="6" height="6" rx="1.5" stroke="currentColor" strokeWidth="1.5" />
    <rect x="10" y="10" width="6" height="6" rx="1.5" stroke="currentColor" strokeWidth="1.5" />
  </svg>
)

export const IconChat = (p: P) => (
  <svg width="18" height="18" viewBox="0 0 18 18" fill="none" {...p}>
    <path d="M15.75 8.5a6.5 6.5 0 0 1-9.5 5.8L2.25 15.5l1.2-4A6.5 6.5 0 1 1 15.75 8.5Z" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
  </svg>
)

export const IconStudents = (p: P) => (
  <svg width="18" height="18" viewBox="0 0 18 18" fill="none" {...p}>
    <path d="M2.25 5.25 9 2.25l6.75 3L9 8.25l-6.75-3Z" stroke="currentColor" strokeWidth="1.5" strokeLinejoin="round" />
    <path d="M5.25 7.5v3.6c0 .8 1.7 1.65 3.75 1.65s3.75-.85 3.75-1.65V7.5" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
  </svg>
)

export const IconLogout = (p: P) => (
  <svg width="18" height="18" viewBox="0 0 18 18" fill="none" {...p}>
    <path d="M7 15.5H4.5a2 2 0 0 1-2-2v-9a2 2 0 0 1 2-2H7" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
    <path d="M11.5 12.5 15 9l-3.5-3.5M15 9H7" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
  </svg>
)

/** Ikon pensil (tombol "Chat baru" siswa). */
export const IconPencil = (p: P) => (
  <svg width="18" height="18" viewBox="0 0 15 15" fill="none" {...p}>
    <path
      d="M12.76 4.34L8.84 8.26c-.79.79-1.18 1.18-1.66 1.45-.47.27-1.6.53-2.68.79.26-1.08.53-2.21.79-2.68.27-.48.66-.87 1.45-1.65L10.66 2.24l1.05-1.06a1.5 1.5 0 0 1 2.1 2.1l-1.05 1.06ZM10.66 2.24l2.1 2.1"
      stroke="currentColor"
      strokeWidth="1.5"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
    <path
      d="M14.25 7.5c0 3.18 0 4.77-.99 5.76-1 1-2.58 1-5.76 1s-4.77 0-5.76-.99C.75 12.27.75 10.68.75 7.5s0-4.77.99-5.76C2.73.75 4.32.75 7.5.75"
      stroke="currentColor"
      strokeWidth="1.5"
      strokeLinecap="round"
    />
  </svg>
)

/** Ikon kartu siswa (dashboard "Daftar Siswa"). */
export const IconStudentCard = (p: P) => (
  <svg width="22" height="19" viewBox="0 0 22 19" fill="none" {...p}>
    <path
      d="M13 1C16.7712 1 18.6569 1 19.8284 2.2448C21 3.48959 21 5.49306 21 9.5C21 13.5069 21 15.5104 19.8284 16.7552C18.6569 18 16.7712 18 13 18L9 18C5.22876 18 3.34315 18 2.17157 16.7552C1 15.5104 1 13.5069 1 9.5C1 5.49306 1 3.48959 2.17158 2.2448C3.34315 1 5.22877 1 9 1L13 1Z"
      stroke="currentColor"
      strokeWidth="1.6"
      strokeLinejoin="round"
    />
    <path
      d="M3.99996 13C5.60861 10.8627 9.35387 10.746 11 13M9.24899 7.75014C9.24899 8.71664 8.46548 9.50014 7.49899 9.50014C6.53249 9.50014 5.74899 8.71664 5.74899 7.75014C5.74899 6.78364 6.53249 6.00014 7.49899 6.00014C8.46548 6.00014 9.24899 6.78364 9.24899 7.75014Z"
      stroke="currentColor"
      strokeWidth="1.6"
      strokeLinejoin="round"
    />
  </svg>
)

/** Ikon idea (insight "Kemampuan & minat"). */
export const IconIdea = (p: P) => (
  <svg width="20" height="20" viewBox="0 0 20 20" fill="none" {...p}>
    <path
      d="M6.66667 13.75C6.14146 13.2767 5.72166 12.6983 5.43453 12.0522C5.1474 11.4062 4.99935 10.707 5 10C5 8.67392 5.52678 7.40215 6.46447 6.46447C7.40215 5.52678 8.67392 5 10 5C11.3261 5 12.5979 5.52678 13.5355 6.46447C14.4732 7.40215 15 8.67392 15 10C15.0006 10.707 14.8526 11.4062 14.5655 12.0522C14.2783 12.6983 13.8585 13.2767 13.3333 13.75"
      stroke="currentColor"
      strokeWidth="1.5"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
    <path
      d="M10 10V13.75M7.91667 15.8333H12.0833M8.75 17.9167H11.25M2.91667 10H2.08333M4.98667 4.98L4.36167 4.35833M15.0167 4.98L15.6417 4.35833M17.9167 10H17.0833M10 2.08333V2.91667"
      stroke="currentColor"
      strokeWidth="1.5"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
  </svg>
)

/** Ikon bintang (insight "Kelebihan"). */
export const IconStar = (p: P) => (
  <svg width="20" height="20" viewBox="0 0 20 20" fill="none" {...p}>
    <path
      d="M11.44 2.87L12.9067 5.8275C13.1067 6.23917 13.64 6.63417 14.09 6.70917L16.7475 7.155C18.4475 7.44083 18.8475 8.68417 17.6225 9.91083L15.5558 11.9942C15.2058 12.3467 15.0142 13.0275 15.1225 13.515L15.7142 16.0942C16.1808 18.1358 15.1058 18.925 13.3142 17.8583L10.8225 16.3708C10.3725 16.1025 9.63083 16.1025 9.1725 16.3708L6.6825 17.8583C4.89917 18.925 3.81583 18.1267 4.2825 16.0942L4.87417 13.515C4.9825 13.0275 4.79083 12.3467 4.44083 11.9942L2.37417 9.91083C1.15833 8.68333 1.55 7.44083 3.24917 7.155L5.9075 6.70917C6.34917 6.63417 6.8825 6.23917 7.0825 5.8275L8.54917 2.87C9.34917 1.26583 10.6492 1.26583 11.44 2.87Z"
      stroke="currentColor"
      strokeWidth="1.5"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
  </svg>
)

/** Ikon hati (insight "Kebutuhan & dukungan"). */
export const IconFavourite = (p: P) => (
  <svg width="20" height="20" viewBox="0 0 20 20" fill="none" {...p}>
    <path
      d="M1.66667 7.245C1.66667 10.8625 6.325 14.8817 8.675 16.64C9.46667 17.2317 10.5333 17.2317 11.325 16.64C13.675 14.8817 18.3333 10.8617 18.3333 7.245C18.3333 4.855 16.5792 2.91667 14.1667 2.91667C12.9167 2.91667 11.6667 3.33333 10 5C8.33333 3.33333 7.08333 2.91667 5.83333 2.91667C3.42083 2.91667 1.66667 4.855 1.66667 7.245Z"
      stroke="currentColor"
      strokeWidth="1.5"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
  </svg>
)

/** Ikon info (insight "Data belum cukup"). */
export const IconInfoCircle = (p: P) => (
  <svg width="20" height="20" viewBox="0 0 20 20" fill="none" {...p}>
    <path
      d="M10 18.3333C14.6024 18.3333 18.3333 14.6024 18.3333 10C18.3333 5.39763 14.6024 1.66667 10 1.66667C5.39763 1.66667 1.66667 5.39763 1.66667 10C1.66667 14.6024 5.39763 18.3333 10 18.3333Z"
      stroke="currentColor"
      strokeWidth="1.5"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
    <path d="M10 13.3333V10M10 6.66667H10.0083" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
  </svg>
)

/** Ikon lampiran (composer). */
export const IconAttach = (p: P) => (
  <svg width="20" height="20" viewBox="0 0 20 20" fill="none" {...p}>
    <path
      d="M11.667 11.25V6.667a3.333 3.333 0 1 0-6.667 0v4.583a5.417 5.417 0 0 0 10.833 0V3.333h1.667v7.917a7.083 7.083 0 0 1-14.167 0V6.667a5 5 0 1 1 10 0v4.583a2.083 2.083 0 1 1-4.166 0V6.667h1.666v4.583a.417.417 0 0 0 .834 0Z"
      fill="currentColor"
    />
  </svg>
)

/** Ikon kirim (composer). */
export const IconSend = (p: P) => (
  <svg width="16" height="16" viewBox="0 0 16 16" fill="none" {...p}>
    <path d="M8.667 14.667H7.333V3.886L4.7 6.519l-.943-.943L8 1.333l4.243 4.243-.943.943-2.633-2.633v10.781Z" fill="currentColor" />
  </svg>
)

/** Ikon user (avatar sidebar, putih di atas gradient). */
export const IconUserLight = (p: P) => (
  <svg width="20" height="20" viewBox="0 0 28 28" fill="none" {...p}>
    <path d="M22.17 24.5v-2.33a4.67 4.67 0 0 0-4.67-4.67h-7A4.67 4.67 0 0 0 5.83 22.17v2.33M18.67 8.17a4.67 4.67 0 1 1-9.34 0 4.67 4.67 0 0 1 9.34 0Z" stroke="#FCFCFE" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
  </svg>
)

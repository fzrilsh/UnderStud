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
      strokeLinecap="round"
    />
    <path d="M14 7L18 7M14 11H16" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" />
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
      d="M11.44 2.87L12.9067 5.8275C13.1067 6.23917 13.64 6.63417 14.09 6.70917L16.7475 7.155C18.4475 7.44083 18.8475 8.68417 17.6225 9.91083L15.5558 11.9942C15.2058 12.3467 15.0142 13.0275 15.1225 13.515L15.7142 16.0942C16.1808 18.1358 15.1058 18.925 13.3142 17.8583L10.8225 16.3708C10.3725 16.1025 9.63083 16.1025 9.1725 16.3708L6.6825 17.8583C4.89917 18.925 3.81583 18.1267 4.2825 16.0942L4.87417 13.515C4.9825 13.0275 4.79083 12.3467 4.44083 11.9942L2.37417 9.91083C1.15833 8.68333 1.55 7.44083 3.24917 7.155L5.9075 6.70917C6.34917 6.63417 6.8825 6.23917 7.0825 5.8275L8.54917 2.87C9.34917 1.26583 10.6492 1.26583 11.44 2.87"
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
    <path d="M10 13.3333V10M10.1042 6.875H10M10.2083 6.875C10.2083 6.81975 10.1864 6.76691 10.1473 6.72782C10.1082 6.68873 10.0554 6.66684 10.0002 6.66684C9.94491 6.66684 9.89207 6.68873 9.85298 6.72782C9.81389 6.76691 9.792 6.81975 9.792 6.875C9.792 6.93025 9.81389 6.98309 9.85298 7.02218C9.89207 7.06127 9.94491 7.08317 10.0002 7.08317C10.0554 7.08317 10.1082 7.06127 10.1473 7.02218C10.1864 6.98309 10.2083 6.93025 10.2083 6.875Z" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
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

/** Ikon meta: user (16px). */
export const IconUserSm = (p: P) => (
  <svg width="16" height="16" viewBox="0 0 16 16" fill="none" {...p}>
    <path
      d="M2.66667 14.6667C2.66667 11.7211 5.05448 9.33333 8 9.33333C10.9455 9.33333 13.3333 11.7211 13.3333 14.6667H12C12 12.4575 10.2091 10.6667 8 10.6667C5.79086 10.6667 4 12.4575 4 14.6667H2.66667ZM8 8.66667C5.79 8.66667 4 6.87667 4 4.66667C4 2.45667 5.79 0.666667 8 0.666667C10.21 0.666667 12 2.45667 12 4.66667C12 6.87667 10.21 8.66667 8 8.66667ZM8 7.33333C9.47333 7.33333 10.6667 6.14 10.6667 4.66667C10.6667 3.19333 9.47333 2 8 2C6.52667 2 5.33333 3.19333 5.33333 4.66667C5.33333 6.14 6.52667 7.33333 8 7.33333Z"
      fill="currentColor"
    />
  </svg>
)

/** Ikon meta: grup (16px). */
export const IconUsersGroup = (p: P) => (
  <svg width="16" height="16" viewBox="0 0 16 16" fill="none" {...p}>
    <path
      d="M10.667 14v-1.333a2.667 2.667 0 0 0-2.667-2.667H4a2.667 2.667 0 0 0-2.667 2.667V14M12.666 14v-1.333a2.667 2.667 0 0 0-2-2.58M6 10a2.667 2.667 0 1 0 0-5.334A2.667 2.667 0 0 0 6 10ZM10.666 4.753a2.667 2.667 0 0 1 0 5.16"
      stroke="currentColor"
      strokeWidth="1.4"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
  </svg>
)

/** Ikon meta: jam (16px). */
export const IconClock = (p: P) => (
  <svg width="16" height="16" viewBox="0 0 16 16" fill="none" {...p}>
    <circle cx="8" cy="8" r="6.333" stroke="currentColor" strokeWidth="1.5" />
    <path d="M8 4.333V8l2.333 1.333" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
  </svg>
)

/** Ikon user (avatar sidebar, putih di atas gradient). */
export const IconUserLight = (p: P) => (
  <svg width="20" height="20" viewBox="0 0 28 28" fill="none" {...p}>
    <path d="M22.17 24.5v-2.33a4.67 4.67 0 0 0-4.67-4.67h-7A4.67 4.67 0 0 0 5.83 22.17v2.33M18.67 8.17a4.67 4.67 0 1 1-9.34 0 4.67 4.67 0 0 1 9.34 0Z" stroke="#FCFCFE" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
  </svg>
)

/** Ikon titik kecil (pemisah meta, 16px). */
export const IconDotSm = (p: P) => (
  <svg width="16" height="16" viewBox="0 0 16 16" fill="none" {...p}>
    <path d="M8.06667 8.73333C8.43486 8.73333 8.73333 8.43486 8.73333 8.06667C8.73333 7.69848 8.43486 7.4 8.06667 7.4C7.69848 7.4 7.4 7.69848 7.4 8.06667C7.4 8.43486 7.69848 8.73333 8.06667 8.73333Z" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
  </svg>
)

/** Ikon titik (pemisah judul header chat, 18px). */
export const IconChatDot = (p: P) => (
  <svg width="18" height="18" viewBox="0 0 18 18" fill="none" {...p}>
    <path d="M9.075 9.825a.75.75 0 1 0 0-1.5.75.75 0 0 0 0 1.5Z" stroke="currentColor" strokeOpacity=".6" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
  </svg>
)

/** Ikon sematkan (18px). */
export const IconChatPin = (p: P) => (
  <svg width="18" height="18" viewBox="0 0 18 18" fill="none" {...p}>
    <path d="M11.25 3.375L8.25 6.375L5.25 7.5L4.125 8.625L9.375 13.875L10.5 12.75L11.625 9.75L14.625 6.75" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
    <path d="M6.75 11.25L3.375 14.625" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
    <path d="M10.875 3L15 7.125" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
  </svg>
)

/** Ikon lepas sematan (18px). */
export const IconChatUnpin = (p: P) => (
  <svg width="18" height="18" viewBox="0 0 18 18" fill="none" {...p}>
    <path d="M11.25 3.375L8.25 6.375L5.25 7.5L4.125 8.625L9.375 13.875L10.5 12.75L11.625 9.75L14.625 6.75" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
    <path d="M6.75 11.25L3.375 14.625" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
    <path d="M10.875 3L15 7.125" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
    <path d="M2.5 15.5L15.5 2.5" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
  </svg>
)

/** Ikon simpan (18px). */
export const IconChatSave = (p: P) => (
  <svg width="18" height="18" viewBox="0 0 18 18" fill="none" {...p}>
    <path d="M12.75 15.75V10.5C12.75 10.3011 12.671 10.1103 12.5303 9.96967C12.3897 9.82902 12.1989 9.75 12 9.75H6C5.80109 9.75 5.61032 9.82902 5.46967 9.96967C5.32902 10.1103 5.25 10.3011 5.25 10.5V15.75M5.25 2.25V5.25C5.25 5.44891 5.32902 5.63968 5.46967 5.78033C5.61032 5.92098 5.80109 6 6 6H11.25M11.4 2.25C11.7957 2.25564 12.1731 2.41738 12.45 2.7L15.3 5.55C15.5826 5.82695 15.7444 6.20435 15.75 6.6V14.25C15.75 14.6478 15.592 15.0294 15.3107 15.3107C15.0294 15.592 14.6478 15.75 14.25 15.75H3.75C3.35218 15.75 2.97064 15.592 2.68934 15.3107C2.40804 15.0294 2.25 14.6478 2.25 14.25V3.75C2.25 3.35218 2.40804 2.97064 2.68934 2.68934C2.97064 2.40804 3.35218 2.25 3.75 2.25H11.4Z" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
  </svg>
)

/** Ikon keluarkan dari draft (18px). */
export const IconChatSaveOff = (p: P) => (
  <svg width="18" height="18" viewBox="0 0 18 18" fill="none" {...p}>
    <path d="M9.75 9.75H5.7C5.48348 9.75 5.27585 9.836 5.12274 9.9891C4.96964 10.1422 4.8836 10.3498 4.8836 10.5664V16.25M10.5 5.75H11.25M12.75 16.25V13M0.75 0.75L16.75 16.75" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
    <path d="M15.3075 15.3075C15.0075 15.6091 14.5984 15.7803 14.1713 15.7822H3.25C2.821 15.7822 2.4096 15.6117 2.1063 15.3084C1.8029 15.005 1.6325 14.5936 1.6325 14.1647V3.24339C1.6343 2.81634 1.8054 2.40719 2.107 2.10722M6.5 1.6322H11.5466C11.9745 1.63823 12.3829 1.81328 12.6825 2.12089L15.7765 5.21489C16.0841 5.51448 16.2592 5.92285 16.2652 6.35072V11.3987" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
  </svg>
)

/** Ikon opsi "…" (18px). */
export const IconChatMore = (p: P) => (
  <svg width="18" height="18" viewBox="0 0 18 18" fill="none" {...p}>
    <circle cx="3.375" cy="9" r="0.9" stroke="currentColor" strokeWidth="1.5" />
    <circle cx="9" cy="9" r="0.9" stroke="currentColor" strokeWidth="1.5" />
    <circle cx="14.625" cy="9" r="0.9" stroke="currentColor" strokeWidth="1.5" />
  </svg>
)

/** Ikon panah kanan terang (18px, tombol "Selesai & lanjut" aktif). */
export const IconChatArrowRight = (p: P) => (
  <svg width="18" height="18" viewBox="0 0 18 18" fill="none" {...p}>
    <path d="M3.75 9H14.25M9 14.25L14.25 9L9 3.75" stroke="#E2E2F2" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
  </svg>
)

/** Ikon panah kanan pudar (18px, tombol "Selesai & lanjut" nonaktif). */
export const IconChatArrowRightMuted = (p: P) => (
  <svg width="18" height="18" viewBox="0 0 18 18" fill="none" {...p}>
    <path d="M3.75 9H14.25M9 14.25L14.25 9L9 3.75" stroke="#71717a" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
  </svg>
)

/** Ikon dokumen terlampir (10px, di dalam bar kartu lampiran). */
export const IconDocAttach = (p: P) => (
  <svg width="10" height="10" viewBox="0 0 7.125 8.25" fill="none" {...p}>
    <path d="M6.75 3.5625V3.375c0-1.4142 0-2.1213-.4393-2.5607C5.8713.375 5.1642.375 3.75.375h-.375c-1.4142 0-2.1213 0-2.5606.4393C.375 1.2537.375 1.9608.375 3.375v1.6875c0 1.2328 0 1.8492.3405 2.2641a1.875 1.875 0 0 0 .208.208c.4148.3405 1.0312.3405 2.264.3405" stroke="currentColor" strokeWidth="0.75" strokeLinecap="round" strokeLinejoin="round" />
  </svg>
)

/** Chevron bawah terisi (18px, grup kelas nav). */
export const IconArrowDownSm = (p: P) => (
  <svg width="18" height="18" viewBox="0 0 18 18" fill="none" {...p}>
    <path d="M9 9.88l3.71-3.71 1.06 1.06L9 12l-4.77-4.77 1.06-1.06L9 9.88Z" fill="currentColor" />
  </svg>
)

/** Ikon buka arsip (16px, unarchive). */
export const IconUnarchive = (p: P) => (
  <svg width="16" height="16" viewBox="0 0 16 16" fill="none" {...p}>
    <rect x="1.9" y="2.6" width="12.2" height="3.2" rx="0.9" stroke="currentColor" strokeWidth="1.3" />
    <path d="M2.9 5.8v5.8A1.4 1.4 0 0 0 4.3 13h7.4a1.4 1.4 0 0 0 1.4-1.4V5.8" stroke="currentColor" strokeWidth="1.3" />
    <path d="M8 11.5V8m0 0L6.6 9.4M8 8l1.4 1.4" stroke="currentColor" strokeWidth="1.3" strokeLinecap="round" strokeLinejoin="round" />
  </svg>
)

/** Ikon tong sampah (15px, hapus chat siswa). */
export const IconTrash = (p: P) => (
  <svg width="15" height="15" viewBox="0 0 15 15" fill="none" {...p}>
    <path d="M1.875 3.75h11.25M6.25 6.875v3.75M8.75 6.875v3.75M2.5 3.75l.625 8.125a1.25 1.25 0 0 0 1.25 1.25h6.25a1.25 1.25 0 0 0 1.25-1.25L12.5 3.75M5.625 3.75V2.5a1.25 1.25 0 0 1 1.25-1.25h1.25a1.25 1.25 0 0 1 1.25 1.25v1.25" stroke="currentColor" strokeWidth="1.3" strokeLinecap="round" strokeLinejoin="round" />
  </svg>
)

/** Ikon kaca pembesar (18px, input pencarian). */
export const IconSearchSm = (p: P) => (
  <svg width="18" height="18" viewBox="0 0 18 18" fill="none" {...p}>
    <path d="M16.5 16.5 12.11 12.11M14.25 7.875a6.375 6.375 0 1 1-12.75 0 6.375 6.375 0 0 1 12.75 0Z" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
  </svg>
)

/** Ikon unggah (16px, "Impor E-Rapor"). */
export const IconUpload = (p: P) => (
  <svg width="16" height="16" viewBox="0 0 16 16" fill="none" {...p}>
    <path d="M8 11.5V3m0 0L5 6m3-3 3 3" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
    <path d="M2 11v2a2 2 0 0 0 2 2h8a2 2 0 0 0 2-2v-2" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
  </svg>
)

/** Ikon user (18px, "Tambah siswa"). */
export const IconUser = (p: P) => (
  <svg width="18" height="18" viewBox="0 0 18 18" fill="none" {...p}>
    <path d="M14.5 15.5v-1.4a3.1 3.1 0 0 0-3.1-3.1H6.6A3.1 3.1 0 0 0 3.5 14.1v1.4M12 5.1a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
  </svg>
)

/** Ikon task-add terang (18px, tombol gelap "Tambah observasi"). */
export const IconTaskAddDark = (p: P) => (
  <svg width="18" height="18" viewBox="0 0 18 18" fill="none" {...p}>
    <path d="M6.2 2.4H4.7A1.35 1.35 0 0 0 3.35 3.75v10.8a1.35 1.35 0 0 0 1.35 1.35h4.85" stroke="#fcfcfe" strokeWidth="1.45" strokeLinecap="round" />
    <rect x="5.6" y="1" width="5.1" height="2.9" rx="1" stroke="#fcfcfe" strokeWidth="1.45" />
    <path d="M6 7.2h3M6 9.8h2" stroke="#fcfcfe" strokeWidth="1.45" strokeLinecap="round" />
    <path d="M13.5 10.8v4.1M11.5 12.8h4.1" stroke="#fcfcfe" strokeWidth="1.55" strokeLinecap="round" />
  </svg>
)

/** Ikon unduh violet (16px, tombol "Unduh" ringkasan). */
export const IconDownloadVioletSm = (p: P) => (
  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" {...p}>
    <path d="M3 17C3 17.93 3 18.395 3.102 18.777C3.23832 19.2856 3.50608 19.7493 3.87839 20.1216C4.2507 20.4939 4.71443 20.7617 5.223 20.898C5.605 21 6.07 21 7 21H17C17.93 21 18.395 21 18.776 20.898C19.2848 20.7618 19.7487 20.4941 20.1212 20.1218C20.4937 19.7495 20.7616 19.2857 20.898 18.777C21 18.395 21 17.93 21 17M16.5 11.5C16.5 11.5 13.186 16 12 16C10.814 16 7.5 11.5 7.5 11.5M12 15V3" stroke="#3D3BA6" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
  </svg>
)

/** Ikon unduh terang (18px, tombol "Unduh PDF" header). */
export const IconDownloadLight = (p: P) => (
  <svg width="18" height="18" viewBox="0 0 18 18" fill="none" {...p}>
    <path d="M2.25 12.75C2.25 13.4475 2.25 13.7963 2.3265 14.0828C2.42874 14.4642 2.62956 14.812 2.90879 15.0912C3.18802 15.3704 3.53582 15.5713 3.91725 15.6735C4.20375 15.75 4.5525 15.75 5.25 15.75H12.75C13.4475 15.75 13.7962 15.75 14.082 15.6735C14.4636 15.5714 14.8115 15.3706 15.0909 15.0914C15.3703 14.8121 15.5712 14.4643 15.6735 14.0828C15.75 13.7963 15.75 13.4475 15.75 12.75M12.375 8.625C12.375 8.625 9.8895 12 9 12C8.1105 12 5.625 8.625 5.625 8.625M9 11.25V2.25" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
  </svg>
)

/** Ikon segarkan (18px, "Buat ulang" ringkasan). */
export const IconRefresh = (p: P) => (
  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" {...p}>
    <path d="M20.01 2V5.132C20.0102 5.19671 19.9904 5.2599 19.9533 5.31293C19.9162 5.36595 19.8636 5.40622 19.8028 5.42822C19.7419 5.45022 19.6757 5.45288 19.6133 5.43584C19.5509 5.41879 19.4952 5.38287 19.454 5.333C18.5167 4.28332 17.3681 3.44372 16.0835 2.8693C14.7988 2.29488 13.4072 1.99863 12 2C6.477 2 2 6.477 2 12C2 17.523 6.477 22 12 22C17.523 22 22 17.523 22 12" stroke="#3D3BA6" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
  </svg>
)

/** Ikon perisai terang (18px, disclaimer ringkasan). */
export const IconShieldLight = (p: P) => (
  <svg width="18" height="18" viewBox="0 0 18 18" fill="none" {...p}>
    <path d="M9 1.5C10.875 1.5 12.6128 1.9155 14.0317 2.62125C14.7277 2.96775 15.075 3.141 15.4125 3.6855C15.75 4.23 15.75 4.7565 15.75 5.811V8.42775C15.75 12.69 12.3427 15.06 10.3702 16.0755L10.3658 16.0778C9.81826 16.359 9.54377 16.5 9 16.5C8.45623 16.5 8.18174 16.359 7.63421 16.0778L7.62975 16.0755C5.6565 15.06 2.25 12.6908 2.25 8.4285V5.811C2.25 4.7565 2.25075 4.23 2.5875 3.6855C2.92425 3.141 3.27225 2.96775 3.96825 2.62125C5.388 1.9155 7.125 1.5 9 1.5Z" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
  </svg>
)

/** Ikon magic AI besar (22px, hero ringkasan). */
export const IconAiMagicHero = (p: P) => (
  <svg width="22" height="22" viewBox="0 0 22 22" fill="none" {...p}>
    <path d="M17.875 1.83333V6.41667M15.5833 4.125H20.1667M12.3081 7.41033C11.7343 5.35608 11.2961 3.66667 10.0833 3.66667C8.87058 3.66667 8.43333 5.35608 7.85858 7.41033C7.62575 8.2445 7.50842 8.6625 7.16833 9.00258C6.82917 9.34175 6.41117 9.45908 5.577 9.69192C3.52275 10.2667 1.83333 10.7039 1.83333 11.9167C1.83333 13.1294 3.52275 13.5667 5.577 14.1414C6.41117 14.3743 6.82917 14.4907 7.16833 14.8308C7.50842 15.1708 7.62575 15.5879 7.85858 16.423C8.43333 18.4773 8.8715 20.1667 10.0833 20.1667C11.2952 20.1667 11.7333 18.4773 12.3081 16.423C12.5418 15.5888 12.6583 15.1708 12.9983 14.8308C13.3375 14.4916 13.7555 14.3743 14.5897 14.1414C16.6439 13.5667 18.3333 13.1294 18.3333 11.9167C18.3333 10.7039 16.6439 10.2667 14.5897 9.69192C13.7555 9.45908 13.3375 9.34175 12.9983 9.00258C12.6583 8.6625 12.5418 8.24542 12.3081 7.41033Z" stroke="#FCFCFE" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
  </svg>
)

/** Ikon magic AI (16px, tombol "Buat" tabel). */
export const IconAiMagic = (p: P) => (
  <svg width="16" height="16" viewBox="0 0 16 16" fill="none" {...p}>
    <path d="M13.0007 1.33301V4.66634M11.334 2.99967H14.6673M8.95198 5.38901C8.53465 3.89501 8.21598 2.66634 7.33398 2.66634C6.45198 2.66634 6.13398 3.89501 5.71598 5.38901C5.54665 5.99567 5.46132 6.29967 5.21398 6.54701C4.96732 6.79367 4.66332 6.87901 4.05665 7.04834C2.56265 7.46634 1.33398 7.78434 1.33398 8.66634C1.33398 9.54834 2.56265 9.86634 4.05665 10.2843C4.66332 10.4537 4.96732 10.5383 5.21398 10.7857C5.46132 11.033 5.54665 11.3363 5.71598 11.9437C6.13398 13.4377 6.45265 14.6663 7.33398 14.6663C8.21532 14.6663 8.53398 13.4377 8.95198 11.9437C9.12198 11.337 9.20665 11.033 9.45398 10.7857C9.70065 10.539 10.0047 10.4537 10.6113 10.2843C12.1053 9.86634 13.334 9.54834 13.334 8.66634C13.334 7.78434 12.1053 7.46634 10.6113 7.04834C10.0047 6.87901 9.70065 6.79367 9.45398 6.54701C9.20665 6.29967 9.12198 5.99634 8.95198 5.38901Z" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
  </svg>
)

/** Ikon magic AI terang (18px, tombol "Simpan & buat ringkasan AI"). */
export const IconAiMagicLight = (p: P) => (
  <svg width="18" height="18" viewBox="0 0 18 18" fill="none" {...p}>
    <path d="M9 1.8l1.5 4.2L14.7 7.5l-4.2 1.5L9 13.2 7.5 9 3.3 7.5 7.5 6 9 1.8Z" stroke="#fcfcfe" strokeWidth="1.4" strokeLinejoin="round" />
    <path d="M14.2 11.4l.7 1.9 1.9.7-1.9.7-.7 1.9-.7-1.9-1.9-.7 1.9-.7.7-1.9Z" fill="#fcfcfe" />
  </svg>
)

/** Ikon anonim (18px, toggle mode lapor). */
export const IconAnon = (p: P) => (
  <svg width="18" height="18" viewBox="0 0 16.5 15" fill="none" {...p}>
    <g stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
      <path d="M4.5 9.75a2.25 2.25 0 1 0 0 4.5 2.25 2.25 0 0 0 0-4.5Z" />
      <path d="M12 9.75a2.25 2.25 0 1 0 0 4.5 2.25 2.25 0 0 0 0-4.5Z" />
      <path d="M9.75 11.25h-3" />
      <path d="M15.75 8.25c-1.84-.92-4.52-1.5-7.5-1.5s-5.66.58-7.5 1.5" />
      <path d="M13.5 7.13l-.79-5.1c-.16-1.03-1.29-1.6-2.2-1.09l-.46.25a3.5 3.5 0 0 1-3.59 0l-.46-.25c-.91-.5-2.04.06-2.2 1.1L3 7.13" />
    </g>
  </svg>
)

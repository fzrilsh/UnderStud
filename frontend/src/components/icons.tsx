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

/** Ikon user (avatar sidebar, putih di atas gradient). */
export const IconUserLight = (p: P) => (
  <svg width="20" height="20" viewBox="0 0 28 28" fill="none" {...p}>
    <path d="M22.17 24.5v-2.33a4.67 4.67 0 0 0-4.67-4.67h-7A4.67 4.67 0 0 0 5.83 22.17v2.33M18.67 8.17a4.67 4.67 0 1 1-9.34 0 4.67 4.67 0 0 1 9.34 0Z" stroke="#FCFCFE" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
  </svg>
)

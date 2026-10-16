import { Outlet } from 'react-router-dom'

/** Layout publik (tanpa sidebar). Halaman mengatur layoutnya sendiri. */
export default function RootLayout() {
  return <Outlet />
}

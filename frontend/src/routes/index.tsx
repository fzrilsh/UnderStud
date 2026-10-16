import { createBrowserRouter } from 'react-router-dom'

import RootLayout from '@/components/layout/root-layout'
import AppLayout from '@/components/layout/app-layout'
import ProtectedRoute from '@/routes/protected-route'
import LoginPage from '@/pages/login'
import DashboardPage from '@/pages/dashboard'
import StudentsPage from '@/pages/students'
import NotFoundPage from '@/pages/not-found'

export const router = createBrowserRouter([
  {
    element: <RootLayout />,
    children: [{ path: '/login', element: <LoginPage /> }],
  },
  {
    element: <ProtectedRoute />,
    children: [
      {
        element: <AppLayout />,
        children: [
          { path: '/', element: <DashboardPage /> },
          { path: '/students', element: <StudentsPage /> },
        ],
      },
    ],
  },
  { path: '*', element: <NotFoundPage /> },
])

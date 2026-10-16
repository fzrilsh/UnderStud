import { createBrowserRouter } from 'react-router-dom'

import RootLayout from '@/components/layout/root-layout'
import AppLayout from '@/components/layout/app-layout'
import RoleGuard from '@/components/role-guard'
import ProtectedRoute from '@/routes/protected-route'
import LoginPage from '@/pages/login'
import DashboardPage from '@/pages/dashboard'
import StudentsPage from '@/pages/students'
import StudentDetailPage from '@/pages/student-detail'
import ReportsPage from '@/pages/reports'
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
          { path: '/students/:id', element: <StudentDetailPage /> },
          // Chat BK: hanya role BK.
          {
            element: <RoleGuard roles={['BK']} />,
            children: [{ path: '/reports', element: <ReportsPage /> }],
          },
        ],
      },
    ],
  },
  { path: '*', element: <NotFoundPage /> },
])

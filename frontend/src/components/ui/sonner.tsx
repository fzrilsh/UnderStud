"use client"

import { Toaster as Sonner, type ToasterProps } from "sonner"

/**
 * Toast bergaya mockup (`.toast` di app.css): pill gelap #15153a, teks putih,
 * radius 10px, 13.5px/500. Varian ok = #0B7A43, err = #B42318. Tanpa ikon.
 */
const Toaster = ({ ...props }: ToasterProps) => {
  return (
    <Sonner
      className="toaster group"
      position="bottom-right"
      duration={2200}
      icons={{
        success: null,
        info: null,
        warning: null,
        error: null,
        loading: null,
      }}
      style={
        {
          "--normal-bg": "#15153a",
          "--normal-text": "#fff",
          "--normal-border": "transparent",
          "--success-bg": "#0B7A43",
          "--success-text": "#fff",
          "--error-bg": "#B42318",
          "--error-text": "#fff",
          "--border-radius": "10px",
        } as React.CSSProperties
      }
      toastOptions={{
        classNames: {
          toast: "cn-toast !text-[13.5px] !font-medium",
        },
      }}
      {...props}
    />
  )
}

export { Toaster }

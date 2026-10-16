import { useEffect } from 'react'

const SITE = 'UnderStud'
const SUFFIX = ` · ${SITE}`

/** Set meta description (buat kalau belum ada). */
function setMeta(name: string, content: string) {
  let el = document.querySelector<HTMLMetaElement>(`meta[name="${name}"]`)
  if (!el) {
    el = document.createElement('meta')
    el.setAttribute('name', name)
    document.head.appendChild(el)
  }
  el.setAttribute('content', content)
}

/**
 * SEO per halaman: set `document.title` + meta description.
 * Judul otomatis diberi sufiks " · UnderStud".
 */
export function useSeo(title: string, description?: string) {
  useEffect(() => {
    document.title = title.endsWith(SITE) ? title : title + SUFFIX
    if (description) setMeta('description', description)
    return () => {
      // tidak mengembalikan title: halaman berikutnya akan menimpanya.
    }
  }, [title, description])
}

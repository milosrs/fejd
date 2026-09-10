import { API_BASE_URL } from "./api"

export function resolveImageUrl(path: string | undefined): string | undefined {
  if (!path) return undefined
  if (/^https?:\/\//.test(path)) return path
  if (/^data:/.test(path)) return path
  return `${API_BASE_URL}${path.startsWith("/") ? path : `/${path}`}`
}

// dataUrlToFile reconstructs a File from a data URL (used to stage unsaved
// images in-memory/localStorage and upload them later).
export function dataUrlToFile(dataUrl: string, name: string): File {
  const [header, base64] = dataUrl.split(",")
  const mime = header.match(/data:(.*?);/)?.[1] ?? "application/octet-stream"
  const binary = atob(base64)
  const bytes = new Uint8Array(binary.length)
  for (let i = 0; i < binary.length; i++) bytes[i] = binary.charCodeAt(i)
  return new File([bytes], name, { type: mime })
}

// fileToDataUrl reads a File into a data URL so an unsaved image can be staged
// locally (shown in the editor preview) before it is uploaded on Save.
export function fileToDataUrl(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(reader.result as string)
    reader.onerror = () => reject(reader.error)
    reader.readAsDataURL(file)
  })
}

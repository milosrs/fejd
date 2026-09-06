import { Capacitor } from "@capacitor/core"
import { Camera, CameraResultType, CameraSource } from "@capacitor/camera"

export type ImageSource = "library" | "camera"

// pickImage returns an image as a File for upload. On native (Capacitor) it
// uses the camera/photo-library plugin; on web it falls back to a file input.
export async function pickImage(source: ImageSource = "library"): Promise<File | null> {
  if (Capacitor.isNativePlatform()) {
    const photo = await Camera.getPhoto({
      resultType: CameraResultType.Base64,
      source: source === "camera" ? CameraSource.Camera : CameraSource.Photos,
      quality: 90,
    })

    const base64 = photo.base64String
    if (!base64) return null

    const mime = `image/${photo.format}`
    const bytes = Uint8Array.from(atob(base64), (c) => c.charCodeAt(0))
    return new File([bytes], `photo.${photo.format}`, { type: mime })
  }

  return new Promise<File | null>((resolve) => {
    const input = document.createElement("input")
    input.type = "file"
    input.accept = "image/*"
    input.onchange = () => resolve(input.files?.[0] ?? null)
    input.click()
  })
}

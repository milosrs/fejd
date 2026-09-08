import { useRef, useState } from "react"
import { QRCodeCanvas } from "qrcode.react"
import { Check, Copy, Download, X } from "lucide-react"
import { Button } from "../ui/button"
import type { Invitation } from "../../hooks/useInvitations"

export function InviteDialog({
  open,
  onClose,
  invitation,
  loading = false,
  error = null,
  filename = "fejd-invite",
}: {
  open: boolean
  onClose: () => void
  invitation: Invitation | null
  loading?: boolean
  error?: string | null
  filename?: string
}) {
  const qrRef = useRef<HTMLDivElement>(null)
  const [copied, setCopied] = useState(false)

  if (!open) return null

  async function copyLink() {
    if (!invitation) return
    try {
      await navigator.clipboard.writeText(invitation.url)
      setCopied(true)
      setTimeout(() => setCopied(false), 1500)
    } catch {
      // Clipboard unavailable (e.g. non-secure context); ignore.
    }
  }

  function downloadQR() {
    const canvas = qrRef.current?.querySelector("canvas")
    if (!canvas) return
    const link = document.createElement("a")
    link.download = `${filename}.png`
    link.href = canvas.toDataURL("image/png")
    link.click()
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4 pt-[calc(1rem+env(safe-area-inset-top))] pb-[calc(1rem+env(safe-area-inset-bottom))]"
      onClick={onClose}
    >
      <div
        className="w-full max-w-sm max-h-[90vh] overflow-y-auto rounded-2xl border border-border bg-background p-6"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-start justify-between">
          <h3 className="text-lg font-semibold text-foreground">Invite link</h3>
          <Button variant="ghost" size="icon-sm" onClick={onClose} aria-label="Close">
            <X className="size-3" />
          </Button>
        </div>

        <p className="mt-2 text-sm text-muted-foreground">
          Share this link or QR code. Scanning it lets someone join this salon as
          an employee.
        </p>

        {loading ? (
          <div className="mt-6 flex items-center justify-center py-12 text-sm text-muted-foreground">
            Generating…
          </div>
        ) : error ? (
          <p className="mt-6 rounded-xl border border-destructive/40 bg-destructive/10 p-3 text-sm text-destructive">
            {error}
          </p>
        ) : invitation ? (
          <div className="mt-6 flex flex-col items-center gap-4">
            <div
              ref={qrRef}
              className="rounded-xl bg-white p-3 ring-1 ring-foreground/10"
            >
              <QRCodeCanvas
                value={invitation.url}
                size={200}
                level="M"
                includeMargin
                bgColor="#ffffff"
                fgColor="#000000"
              />
            </div>

            <div className="w-full rounded-xl border border-border bg-muted p-3">
              <p className="break-all text-xs text-muted-foreground">{invitation.url}</p>
            </div>

            <div className="flex w-full gap-2">
              <Button className="flex-1" onClick={copyLink}>
                {copied ? <Check className="size-3" /> : <Copy className="size-3" />}
                {copied ? "Copied" : "Copy link"}
              </Button>
              <Button variant="outline" className="flex-1" onClick={downloadQR}>
                <Download className="size-3" />
                Download QR
              </Button>
            </div>

            <p className="text-xs text-muted-foreground">
              Single use · expires {new Date(invitation.expires_at).toLocaleString()}
            </p>
          </div>
        ) : null}
      </div>
    </div>
  )
}

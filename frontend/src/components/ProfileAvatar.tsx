import { useEffect, useState } from "react"
import { UserRound } from "lucide-react"

export function ProfileAvatar({ src, name }: { src?: string; name?: string }) {
  const [loaded, setLoaded] = useState(false)
  const [failed, setFailed] = useState(false)

  useEffect(() => {
    setLoaded(false)
    setFailed(false)
  }, [src])

  const showImage = !!src && !failed

  return (
    <span className="relative flex h-9 w-9 items-center justify-center overflow-hidden rounded-full bg-muted text-muted-foreground">
      {showImage && (
        <>
          <img
            src={src}
            alt={name}
            onLoad={() => setLoaded(true)}
            onError={() => setFailed(true)}
            className={`absolute inset-0 h-full w-full object-cover transition-opacity ${loaded ? "opacity-100" : "opacity-0"
              }`}
          />
          {!loaded && (
            <span className="absolute inset-0 animate-pulse rounded-full bg-muted" />
          )}
        </>
      )}
      {!showImage && <UserRound className="size-5" />}
    </span>
  )
}

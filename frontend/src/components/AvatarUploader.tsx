'use client'

import { useState } from 'react'
import Image from 'next/image'

export default function AvatarUploader({
  currentAvatar,
  onChange,
}: {
  currentAvatar: string
  onChange: (file: File) => void
}) {
  const [preview, setPreview] = useState(currentAvatar)

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (file) {
      setPreview(URL.createObjectURL(file))
      onChange(file)
    }
  }

  return (
    <div className="flex items-center gap-4">
      <div className="relative w-20 h-20 rounded-full overflow-hidden border">
        <Image src={preview} alt="Profile" fill className="object-cover" />
      </div>
      <input type="file" accept="image/*" onChange={handleFileChange} />
    </div>
  )
}

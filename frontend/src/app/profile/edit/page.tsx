'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'

export default function EditProfilePage() {
  const router = useRouter()
  const [name, setName] = useState('John Doe')
  const [username, setUsername] = useState('johnny123')
  const [email, setEmail] = useState('john@example.com')
  const [bio, setBio] = useState('CS student who loves algorithms and memes')

  return (
    <div className="max-w-2xl mx-auto mt-10 px-4">
      <h1 className="text-3xl font-bold mb-6">Edit Profile</h1>
      <form
        onSubmit={e => {
          e.preventDefault()
          // Replace with API call
          console.log({ name, username, email, bio })
          router.push('/profile')
        }}
        className="space-y-4"
      >
        <input
          type="text"
          placeholder="Full Name"
          value={name}
          onChange={e => setName(e.target.value)}
          className="w-full border p-2 rounded"
        />
        <input
          type="text"
          placeholder="Username"
          value={username}
          onChange={e => setUsername(e.target.value)}
          className="w-full border p-2 rounded"
        />
        <input
          type="email"
          placeholder="Email"
          value={email}
          onChange={e => setEmail(e.target.value)}
          className="w-full border p-2 rounded"
        />
        <textarea
          placeholder="Bio"
          value={bio}
          onChange={e => setBio(e.target.value)}
          className="w-full border p-2 rounded resize-none"
        />
        <div className="flex justify-end gap-2">
          <button
            type="button"
            onClick={() => router.back()}
            className="px-4 py-2 rounded border"
          >
            Cancel
          </button>
          <button
            type="submit"
            className="px-4 py-2 rounded bg-blue-600 text-white hover:bg-blue-700"
          >
            Save Changes
          </button>
        </div>
      </form>
    </div>
  )
}

import Image from 'next/image'

export default function ProfileCard({
  name,
  username,
  email,
  bio,
  avatarUrl,
}: {
  name: string
  username: string
  email: string
  bio?: string
  avatarUrl: string
}) {
  return (
    <div className="bg-white rounded-xl shadow p-6 space-y-4 flex items-center gap-6">
      <div className="relative w-24 h-24 rounded-full overflow-hidden border">
        <Image src={avatarUrl} alt="Avatar" fill className="object-cover" />
      </div>
      <div>
        <h2 className="text-2xl font-semibold">{name}</h2>
        <p className="text-gray-600">@{username}</p>
        <p className="text-gray-800">{email}</p>
        {bio && <p className="text-gray-500 italic">{bio}</p>}
      </div>
    </div>
  )
}

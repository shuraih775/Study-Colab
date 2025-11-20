import Link from 'next/link'
import ProfileCard from '@/components/ProfileCard'

export default function ProfilePage() {
  const dummyProfile = {
    name: 'John Doe',
    username: 'johnny123',
    email: 'john@example.com',
    bio: 'CS student who loves algorithms and memes',
    avatarUrl: ''
  }

  return (
    <div className="max-w-2xl mx-auto mt-10 px-4">
      <div className="flex justify-between items-center mb-4">
        <h1 className="text-3xl font-bold">My Profile</h1>
        <Link href="/profile/edit" className="text-blue-600 hover:underline">Edit Profile</Link>
      </div>
      <ProfileCard {...dummyProfile} />
    </div>
  )
}

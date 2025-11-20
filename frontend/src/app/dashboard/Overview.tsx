'use client'

import { useState, useEffect } from 'react'
import CreateGroupModal from '@/components/Modals/CreateGroupModal'
import JoinPublicGroupModal from '@/components/Modals/JoinPublicGroupModal'
import { useRouter } from 'next/navigation'
import { useUser } from '@/context/UserProvider'


export default function Overview() {
  const router = useRouter();
  const [showCreate, setShowCreate] = useState(false)
  const [showJoinPublic, setShowJoinPublic] = useState(false)
  const [showJoinPrivate, setShowJoinPrivate] = useState(false)
  const [isClient, setIsClient] = useState(false)
  
  
  const [now, setNow] = useState(new Date())
  const user = useUser();

  useEffect(() => {
  const interval = setInterval(() => {
    setNow(new Date())
  }, 60000) // updates every minute
  return () => clearInterval(interval)
}, [])
  


  return (
    <div className="min-h-screen  p-6">
      <header className="mb-6">
        <h1 className="text-3xl font-bold text-gray-800">Welcome back, {user?.username} 👋</h1>
        <p className="text-gray-500 mt-1">Here's a quick overview of your study groups.</p>
      </header>

      <section className="grid gap-6 grid-cols-1 md:grid-cols-2">
        <div className="bg-white p-6 rounded-xl shadow-md">
          <h2 className="text-xl font-semibold mb-4">Quick Actions</h2>
          <div className="flex flex-col gap-3">
            <button onClick={() => setShowCreate(true)} className="bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700 transition cursor-pointer">
              Create a Group
            </button>
            <button onClick={() => router.push('/groups')} className="bg-green-600 text-white px-4 py-2 rounded-md hover:bg-green-700 transition cursor-pointer">
              Join a Public Group
            </button>
            
          </div>
        </div>

        
      </section>

      {/* <section className="mt-8 bg-white p-6 rounded-xl shadow-md">
        <h2 className="text-xl font-semibold mb-4">Recent Activity</h2>
        <p className="text-gray-500 text-sm">Coming soon — chats, uploads, and tasks will appear here.</p>
      </section> */}

      
      <CreateGroupModal open={showCreate} onClose={() => setShowCreate(false)} />
      <JoinPublicGroupModal open={showJoinPublic} onClose={() => setShowJoinPublic(false)} />

    </div>
  )
}

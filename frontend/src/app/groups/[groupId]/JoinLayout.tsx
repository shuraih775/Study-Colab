'use client'

import { useState } from 'react'
import axios from 'axios'
import { useRouter } from 'next/navigation'
import { Loader2, AlertCircle, LogIn, Users, Lock } from 'lucide-react' // Added icons

export default function JoinLayout({ groupId, token, groupInfo }: { groupId: string, token?: string, groupInfo?: any }) {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const router = useRouter()

  const joinGroup = async () => {
    setLoading(true)
    setError('')
    try {
      await axios.post(`http://localhost:8080/groups/${groupId}/join`, {}, {
        withCredentials: true
      })
      router.refresh() // Refresh to re-fetch membership status
    } catch (err: any) {
      console.error(err)
      setError(err.response?.data?.message || 'Failed to send join request.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="flex min-h-screen justify-center items-center bg-gray-50 p-4">
      <div className="bg-white p-8 rounded-xl shadow-lg w-full max-w-lg text-center">
        {/* Group Header */}
        <div className="mb-6">
          <div className="mx-auto flex items-center justify-center h-16 w-16 rounded-full bg-emerald-100 mb-4">
            <LogIn className="h-8 w-8 text-emerald-600" />
          </div>
          <h2 className="text-2xl font-bold text-gray-800 mb-2">
            Join {groupInfo?.name || 'This Group'}
          </h2>
          {groupInfo?.type && (
            <span className={`inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-xs font-medium ${
              groupInfo.type === 'private' 
                ? 'bg-amber-100 text-amber-700' 
                : 'bg-emerald-100 text-emerald-700'
            }`}>
              {groupInfo.type === 'private' ? <Lock size={12} /> : <Users size={12} />}
              {groupInfo.type}
            </span>
          )}
        </div>
        
        <p className="mb-6 text-gray-600">
          {groupInfo?.description || 'You are not a member of this group yet. Join to start participating!'}
        </p>

        {error && (
          <div className="bg-red-100 border border-red-300 text-red-700 px-4 py-3 rounded-lg flex items-center gap-2 mb-4">
            <AlertCircle size={18} />
            <span className="text-sm">{error}</span>
          </div>
        )}
        
        <button
          onClick={joinGroup}
          disabled={loading}
          className={`
            w-full px-4 py-3 rounded-lg font-semibold text-white transition-all duration-300
            flex items-center justify-center gap-2
            ${loading 
              ? 'bg-gray-400 cursor-not-allowed' 
              : 'bg-emerald-500 hover:bg-emerald-600 focus:outline-none focus:ring-2 focus:ring-emerald-400 focus:ring-opacity-75'
            }
          `}
        >
          {loading ? (
            <>
              <Loader2 className="animate-spin" size={20} />
              Processing...
            </>
          ) : (
            'Send Join Request'
          )}
        </button>
      </div>
    </div>
  )
}

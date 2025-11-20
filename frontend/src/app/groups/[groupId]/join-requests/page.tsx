'use client'

import { useState, useEffect, use, useCallback } from 'react'
import { Check, X, User, Clock, Inbox, Loader2, AlertCircle } from 'lucide-react'
import { format, formatDistanceToNow } from 'date-fns'
import axios from 'axios'

export default function GroupJoinRequestsPage({params}: {params: {groupId: string}}) {
  const { groupId } = use(params);

  const [requests, setRequests] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const getRequests = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const res = await axios.get(`http://localhost:8080/groups/${groupId}/requests`, {
        withCredentials: true,
      })
      // Assuming res.data is the array of requests
      setRequests(res.data || [])
    } catch (err) {
      console.error(err)
      setError("Failed to fetch join requests.")
    } finally {
      setLoading(false)
    }
  }, [groupId])

  useEffect(() => {
    getRequests()
  }, [getRequests])

  const handleRequest = async (userId: string, action: 'accept' | 'reject') => {
    // Optimistically remove from list for faster UI feedback
    setRequests(currentRequests => currentRequests.filter(req => req.id !== userId))
    
    try {
      if (action === 'accept') {
        // This is an UPDATE action, setting status to 'active'
        await axios.put(
          `http://localhost:8080/groups/${groupId}/requests/${userId}`,
          { status: 'active' }, // Send body to update status
          { withCredentials: true }
        )
      } else {
        // This is a DELETE action
        await axios.delete(
          `http://localhost:8080/groups/${groupId}/requests/${userId}`,
          { withCredentials: true }
        )
      }
      // Optional: Re-fetch to ensure consistency after action
      // getRequests(); 
      // Note: We are already optimistically removing. If the API call fails,
      // we should add the request back or show an error toast.
      // For simplicity, we'll just log the error if it fails.
      
    } catch (err) {
      console.error(`Failed to ${action} request:`, err)
      // If the API call fails, re-fetch to get the correct state
      getRequests()
      // You could also show a toast notification here
    }
  }

  const renderContent = () => {
    if (loading) {
      return (
        <div className="flex justify-center items-center h-64">
          <Loader2 className="animate-spin text-purple-600" size={48} />
        </div>
      )
    }

    if (error) {
      return (
        <div className="text-center text-red-700 mt-16 p-6 bg-red-100 rounded-lg">
          <AlertCircle size={48} className="mx-auto mb-4" />
          <h3 className="text-xl font-semibold mb-2">Error</h3>
          <p>{error}</p>
        </div>
      )
    }

    if (requests.length === 0) {
      return (
        <div className="text-center text-gray-500 mt-16 p-6 bg-gray-50 rounded-lg">
          <Inbox size={48} className="mx-auto text-gray-400 mb-4" />
          <h3 className="text-xl font-semibold mb-2">Inbox Clear</h3>
          <p>There are no pending join requests.</p>
        </div>
      )
    }

    return (
      <div className="bg-white rounded-xl shadow-lg border overflow-hidden">
        <ul className="divide-y divide-gray-200">
          {requests.map((req) => (
            <li key={req.id} className="p-4 sm:p-6 flex flex-col sm:flex-row justify-between items-start sm:items-center">
              <div className="flex items-center gap-4 mb-4 sm:mb-0">
                {/* Avatar Placeholder */}
                <div className="flex-shrink-0 w-12 h-12 rounded-full bg-purple-100 flex items-center justify-center border-2 border-purple-200">
                  <span className="text-xl font-semibold text-purple-700">
                    {req.user ? req.user.split(' ').map(n => n[0]).join('').substring(0,2) : <User />}
                  </span>
                </div>
                {/* User Info */}
                <div>
                  <div className="font-semibold text-gray-800">{req.user}</div>
                  <div className="text-sm text-gray-500">{req.email}</div>
                  <div className="text-sm text-gray-400 flex items-center gap-1 mt-1">
                    <Clock size={14} />
                    {/* <span>
                      Requested {formatDistanceToNow(new Date(req.requestedAt), { addSuffix: true })}
                    </span> */}
                  </div>
                </div>
              </div>
              {/* Action Buttons */}
              <div className="flex gap-3 w-full sm:w-auto">
                <button
                  onClick={() => handleRequest(req.id, 'reject')}
                  className="flex-1 sm:flex-none px-4 py-2 bg-red-100 text-red-700 rounded-lg font-semibold transition-colors duration-300 hover:bg-red-200 flex items-center justify-center gap-1.5"
                >
                  <X size={16} />
                  <span>Reject</span>
                </button>
                <button
                  onClick={() => handleRequest(req.id, 'accept')}
                  className="flex-1 sm:flex-none px-4 py-2 bg-emerald-500 text-white rounded-lg font-semibold transition-colors duration-300 hover:bg-emerald-600 flex items-center justify-center gap-1.5"
                >
                  <Check size={16} />
                  <span>Accept</span>
                </button>
              </div>
            </li>
          ))}
        </ul>
      </div>
    )
  }

  return (
    <div className="max-w-4xl mx-auto">
      <h1 className="text-3xl font-bold text-gray-800 mb-8">Join Requests</h1>
      {renderContent()}
    </div>
  )
}

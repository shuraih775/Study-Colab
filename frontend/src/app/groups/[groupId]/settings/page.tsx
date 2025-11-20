'use client'

import { useGroup } from '../GroupContext'
import axios from 'axios'
import { useState, useEffect } from 'react'
import { Loader2, Check, AlertCircle } from 'lucide-react'

type Status = 'idle' | 'loading' | 'success' | 'error'

export default function GroupSettingsPage() {
  const { group, setGroup } = useGroup()
  
  const [name, setName] = useState('')
  const [type, setType] = useState<'public' | 'private'>('public')
  const [description, setDescription] = useState('')
  const [status, setStatus] = useState<Status>('idle')
  const [errorMessage, setErrorMessage] = useState('')

  useEffect(() => {
    if (group) {
      setName(group.name || '')
      setType(group.type || 'public')
      setDescription(group.description || '')
    }
  }, [group])

  if (!group) {
    return (
      <div className="flex justify-center items-center h-64">
        <Loader2 className="animate-spin text-emerald-500" size={48} />
      </div>
    )
  }

  const updateGroup = async () => {
    setStatus('loading')
    setErrorMessage('')
    try {
      const res = await axios.put(`http://localhost:8080/groups/${group.id}`, { // Fixed: use group.id
        name,
        type,
        description,
      }, {
        withCredentials: true
      })
      
      // Update context with new group data
      setGroup(res.data.group)
      setStatus('success')
      setTimeout(() => setStatus('idle'), 2000) // Reset status
    } catch (error: any) {
      console.error('Failed to update group:', error)
      setErrorMessage(error.response?.data?.message || 'Error updating group')
      setStatus('error')
    }
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    updateGroup()
  }

  return (
    <div className="max-w-2xl mx-auto">
      <h1 className="text-3xl font-bold text-gray-800 mb-8">Group Settings</h1>
      <form onSubmit={handleSubmit} className="space-y-6 bg-white p-8 rounded-xl shadow-lg border">
        
        {/* Form Input: Name */}
        <div>
          <label htmlFor="group-name" className="block text-sm font-medium text-gray-700 mb-1">
            Group Name
          </label>
          <input  
            id="group-name"
            type="text"  
            className="w-full border border-gray-300 p-3 rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-purple-500"  
            value={name}  
            onChange={(e) => setName(e.target.value)}  
            required
          />
        </div>

        {/* Form Input: Type */}
        <div>
          <label htmlFor="group-type" className="block text-sm font-medium text-gray-700 mb-1">
            Group Type
          </label>
          <select  
            id="group-type"
            className="w-full border border-gray-300 p-3 rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-purple-500 bg-white"  
            value={type}  
            onChange={(e) => setType(e.target.value as 'public' | 'private')}
            required
          >
            <option value="public">Public (Anyone can join)</option>
            <option value="private">Private (Requires admin approval)</option>
          </select>
        </div>

        {/* Form Input: Description */}
        <div>
          <label htmlFor="group-desc" className="block text-sm font-medium text-gray-700 mb-1">
            Description
          </label>
          <textarea  
            id="group-desc"
            className="w-full border border-gray-300 p-3 rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-purple-500"  
            value={description}  
            onChange={(e) => setDescription(e.target.value)}
            rows={4}
            placeholder="What is this group about?"
          />
        </div>
        
        {/* Error Message */}
        {status === 'error' && (
          <div className="bg-red-100 border border-red-300 text-red-700 px-4 py-3 rounded-lg flex items-center gap-2">
            <AlertCircle size={18} />
            <span className="text-sm">{errorMessage}</span>
          </div>
        )}

        {/* Submit Button */}
        <button  
          type="submit"
          disabled={status === 'loading'}
          className={`
            w-full px-4 py-3 rounded-lg font-semibold text-white transition-all duration-300
            flex items-center justify-center gap-2
            ${status === 'loading' ? 'bg-gray-400 cursor-not-allowed' : ''}
            ${status === 'idle' ? 'bg-emerald-500 hover:bg-emerald-600' : ''}
            ${status === 'success' ? 'bg-green-500' : ''}
            ${status === 'error' ? 'bg-emerald-500 hover:bg-emerald-600' : ''}
          `}
        >
          {status === 'loading' && <Loader2 className="animate-spin" size={20} />}
          {status === 'success' && <Check size={20} />}
          {status === 'idle' && 'Save Changes'}
          {status === 'error' && 'Retry Save'}
        </button>
      </form>
    </div>
  )
}

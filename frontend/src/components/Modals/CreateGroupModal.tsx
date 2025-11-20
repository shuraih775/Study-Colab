'use client'

import { useState, Fragment } from 'react'
import axios from 'axios'
import { useRouter } from 'next/navigation'
import { 
  X, 
  Loader2, 
  Users, 
  Lock, 
  AlertCircle 
} from 'lucide-react'
import { Transition, Dialog } from '@headlessui/react' // Using Headless UI for an accessible modal

/**
 * A beautified modal component for creating a new group.
 * Includes loading, error handling, and better UI elements.
 */
export default function CreateGroupModal({ open, onClose }: { open: boolean; onClose: () => void }) {
  const router = useRouter()
  const [groupName, setGroupName] = useState('')
  const [groupDescription, setGroupDescription] = useState('')
  const [isPrivate, setIsPrivate] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleClose = () => {
    if (isLoading) return // Don't close while loading
    // Reset form on close
    setGroupName('')
    setGroupDescription('')
    setIsPrivate(false)
    setError(null)
    onClose()
  }

  const handleGroupCreation = async () => {
    if (!groupName.trim()) {
      setError("Group name is required.")
      return
    }
    
    setIsLoading(true)
    setError(null)

    try {
      const res = await axios.post('http://localhost:8080/groups',
        {
          name: groupName,
          description: groupDescription,
          type: isPrivate ? "private" : "public"
        },
        {
          withCredentials: true,
        }
      )
      
      const dest = `/groups/${res.data.group_id}/chat`
      router.push(dest)
      handleClose() // Close and reset *after* successful navigation
    } catch (err) {
      console.error(err)
      setError("Failed to create group. Please try again.")
    } finally {
      setIsLoading(false)
    }
  }
  
  // Disable create button if loading or no name is provided
  const isCreateDisabled = isLoading || !groupName.trim()

  return (
    <Transition appear show={open} as={Fragment}>
      <Dialog as="div" className="relative z-50" onClose={handleClose}>
        {/* Backdrop overlay */}
        <Transition.Child
          as={Fragment}
          enter="ease-out duration-300"
          enterFrom="opacity-0"
          enterTo="opacity-100"
          leave="ease-in duration-200"
          leaveFrom="opacity-100"
          leaveTo="opacity-0"
        >
          <div className="fixed inset-0 bg-black/50 backdrop-blur-sm" />
        </Transition.Child>

        <div className="fixed inset-0 overflow-y-auto">
          <div className="flex min-h-full items-center justify-center p-4 text-center">
            {/* Modal panel */}
            <Transition.Child
              as={Fragment}
              enter="ease-out duration-300"
              enterFrom="opacity-0 scale-95"
              enterTo="opacity-100 scale-100"
              leave="ease-in duration-200"
              leaveFrom="opacity-100 scale-100"
              leaveTo="opacity-0 scale-95"
            >
              <Dialog.Panel className="w-full max-w-md transform overflow-hidden rounded-2xl bg-white p-6 text-left align-middle shadow-xl transition-all">
                {/* Header */}
                <div className="flex items-center justify-between mb-4">
                  <Dialog.Title
                    as="h3"
                    className="text-lg font-medium leading-6 text-gray-900"
                  >
                    Create a New Group
                  </Dialog.Title>
                  <button
                    onClick={handleClose}
                    className="p-1 rounded-full text-gray-400 hover:bg-gray-100 hover:text-gray-600 focus:outline-none focus:ring-2 focus:ring-amber-500"
                    aria-label="Close modal"
                    disabled={isLoading}
                  >
                    <X size={20} />
                  </button>
                </div>

                {/* Form */}
                <div className="space-y-4">
                  {/* Group Name Input */}
                  <div>
                    <label htmlFor="groupName" className="block text-sm font-medium text-gray-700 mb-1">
                      Group Name <span className="text-red-500">*</span>
                    </label>
                    <input
                      id="groupName"
                      type="text"
                      placeholder="My Awesome Group"
                      value={groupName}
                      onChange={e => setGroupName(e.target.value)}
                      className="w-full px-4 py-2 border border-gray-300 rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-amber-500 focus:border-amber-500"
                      disabled={isLoading}
                    />
                  </div>

                  {/* Group Description Input */}
                  <div>
                    <label htmlFor="groupDescription" className="block text-sm font-medium text-gray-700 mb-1">
                      Description (Optional)
                    </label>
                    <textarea
                      id="groupDescription"
                      placeholder="What is this group about?"
                      value={groupDescription}
                      onChange={e => setGroupDescription(e.target.value)}
                      rows={3}
                      className="w-full px-4 py-2 border border-gray-300 rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-amber-500 focus:border-amber-500 resize-none"
                      disabled={isLoading}
                    />
                  </div>

                  {/* Private Group Toggle */}
                  <div
                    onClick={() => !isLoading && setIsPrivate(!isPrivate)}
                    className={`flex items-center justify-between p-3 border rounded-lg cursor-pointer ${isPrivate ? 'bg-amber-50 border-amber-300' : 'bg-gray-50 border-gray-200'} ${isLoading ? 'opacity-50 cursor-not-allowed' : 'hover:border-amber-400'}`}
                  >
                    <div className="flex items-center gap-3">
                      <div className={`p-1.5 rounded-full ${isPrivate ? 'bg-amber-600 text-white' : 'bg-gray-300 text-gray-600'}`}>
                        {isPrivate ? <Lock size={16} /> : <Users size={16} />}
                      </div>
                      <div>
                        <p className="font-medium text-gray-800">{isPrivate ? 'Private Group' : 'Public Group'}</p>
                        <p className="text-xs text-gray-500">
                          {isPrivate ? 'Join request should be accepted to become a member.' : 'Anyone can find and join.'}
                        </p>
                      </div>
                    </div>
                    {/* Custom Toggle Switch */}
                    <div className={`relative inline-flex items-center h-6 rounded-full w-11 transition-colors ${isPrivate ? 'bg-amber-600' : 'bg-gray-300'}`}>
                      <span className={`inline-block w-4 h-4 transform transition-transform bg-white rounded-full ${isPrivate ? 'translate-x-6' : 'translate-x-1'}`} />
                    </div>
                  </div>

                  {/* Error Message */}
                  {error && (
                    <div className="flex items-center gap-2 p-3 text-sm text-red-700 bg-red-100 rounded-lg">
                      <AlertCircle size={16} />
                      <span>{error}</span>
                    </div>
                  )}
                </div>

                {/* Footer Buttons */}
                <div className="flex justify-end gap-3 mt-6">
                  <button
                    type="button"
                    onClick={handleClose}
                    className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg shadow-sm hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-amber-500 disabled:opacity-50"
                    disabled={isLoading}
                  >
                    Cancel
                  </button>
                  <button
                    type="button"
                    onClick={handleGroupCreation}
                    className="flex items-center justify-center gap-2 px-4 py-2 text-sm font-medium text-white bg-amber-600 border border-transparent rounded-lg shadow-sm hover:bg-amber-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-amber-500 disabled:opacity-50 disabled:cursor-not-allowed"
                    disabled={isCreateDisabled}
                  >
                    {isLoading ? (
                      <>
                        <Loader2 size={16} className="animate-spin" />
                        Creating...
                      </>
                    ) : (
                      'Create Group'
                    )}
                  </button>
                </div>
              </Dialog.Panel>
            </Transition.Child>
          </div>
        </div>
      </Dialog>
    </Transition>
  )
}

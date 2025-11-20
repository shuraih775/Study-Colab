'use client'

import { useState, Fragment } from 'react'
import axios from 'axios'
import { 
  X, 
  Loader2, 
  AlertCircle,
  Calendar,
  ClipboardList,
  Tag,
  AlignLeft
} from 'lucide-react'
import { Transition, DialogPanel , Dialog} from '@headlessui/react' // Using Headless UI for an accessible modal

/**
 * A beautified modal component for creating a new task within a group.
 * Includes loading, error handling, and better UI elements.
 */
export default function CreateTaskModal({ open, onClose, groupId }: { open: boolean; onClose: () => void; groupId: string }) {
  
  // Helper to get current time in YYYY-MM-DDTHH:MM format
  const getNow = () => new Date().toISOString().slice(0, 16);

  const [title, setTitle] = useState('')
  const [status, setStatus] = useState('pending') // Default status
  const [description, setDescription] = useState('')
  const [deadline, setDeadline] = useState(getNow())
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleClose = () => {
    if (isLoading) return // Don't close while loading
    // Reset form on close
    setTitle('')
    setDescription('')
    setStatus('pending')
    setDeadline(getNow())
    setError(null)
    onClose()
  }

  const handleTaskCreation = async () => {
    if (!title.trim()) {
      setError("Task title is required.")
      return
    }
    
    setIsLoading(true)
    setError(null)

    try {
      await axios.post(`http://localhost:8080/groups/${groupId}/tasks`,
        {
          title,
          description,
          status,
          deadline // Make sure backend can parse this ISO string
        },
        {
          withCredentials: true,
        }
      )
      
      // On success, just close the modal. 
      // The parent component should be responsible for refreshing the task list.
      handleClose()
    } catch (err) {
      console.error(err)
      setError("Failed to create task. Please try again.")
    } finally {
      setIsLoading(false)
    }
  }
  
  // Disable create button if loading or no title is provided
  const isCreateDisabled = isLoading || !title.trim()

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
              <DialogPanel className="w-full max-w-md transform overflow-hidden rounded-2xl bg-white p-6 text-left align-middle shadow-xl transition-all">
                {/* Header */}
                <div className="flex items-center justify-between mb-6">
                  <Dialog.Title
                    as="h3"
                    className="text-lg font-medium leading-6 text-gray-900 flex items-center gap-2"
                  >
                    <ClipboardList size={20} className="text-amber-600" />
                    Create New Task
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
                  {/* Task Title Input */}
                  <div>
                    <label htmlFor="taskTitle" className="block text-sm font-medium text-gray-700 mb-1">
                      Title <span className="text-red-500">*</span>
                    </label>
                    <input
                      id="taskTitle"
                      type="text"
                      placeholder="e.g., Finalize project proposal"
                      value={title}
                      onChange={e => setTitle(e.target.value)}
                      className="w-full px-4 py-2 border border-gray-300 rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-amber-500 focus:border-amber-500"
                      disabled={isLoading}
                    />
                  </div>
                  
                  {/* Status & Deadline Row */}
                  <div className="flex flex-col sm:flex-row gap-4">
                    {/* Status Select */}
                    <div className="flex-1">
                      <label htmlFor="taskStatus" className="block text-sm font-medium text-gray-700 mb-1">
                        Status
                      </label>
                      <div className="relative">
                        <Tag size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
                        <select
                          id="taskStatus"
                          value={status}
                          onChange={e => setStatus(e.target.value)}
                          className="w-full pl-9 pr-4 py-2 border border-gray-300 rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-amber-500 focus:border-amber-500 appearance-none"
                          disabled={isLoading}
                        >
                          <option value="pending">Pending</option>
                          <option value="in_progress">In Progress</option>
                          <option value="completed">Completed</option>
                        </select>
                      </div>
                    </div>
                    
                    {/* Deadline Input */}
                    <div className="flex-1">
                      <label htmlFor="taskDeadline" className="block text-sm font-medium text-gray-700 mb-1">
                        Deadline
                      </label>
                      <div className="relative">
                        <Calendar size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
                        <input
                          id="taskDeadline"
                          type="datetime-local"
                          value={deadline}
                          onChange={(e) => setDeadline(e.target.value)}
                          className="w-full pl-9 pr-2 py-2 border border-gray-300 rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-amber-500 focus:border-amber-500"
                          min={getNow()}
                          disabled={isLoading}
                        />
                      </div>
                    </div>
                  </div>

                  {/* Task Description Input */}
                  <div>
                    <label htmlFor="taskDescription" className="block text-sm font-medium text-gray-700 mb-1">
                      Description (Optional)
                    </label>
                    <div className="relative">
                      <AlignLeft size={16} className="absolute left-3 top-3.5 text-gray-400" />
                      <textarea
                        id="taskDescription"
                        placeholder="Add more details about the task..."
                        value={description}
                        onChange={e => setDescription(e.target.value)}
                        rows={3}
                        className="w-full pl-9 px-4 py-2 border border-gray-300 rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-amber-500 focus:border-amber-500 resize-none"
                        disabled={isLoading}
                      />
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
                    onClick={handleTaskCreation}
                    className="flex items-center justify-center gap-2 px-4 py-2 text-sm font-medium text-white bg-amber-600 border border-transparent rounded-lg shadow-sm hover:bg-amber-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-amber-500 disabled:opacity-50 disabled:cursor-not-allowed"
                    disabled={isCreateDisabled}
                  >
                    {isLoading ? (
                      <>
                        <Loader2 size={16} className="animate-spin" />
                        Creating...
                      </>
                    ) : (
                      'Create Task'
                    )}
                  </button>
                </div>
              </DialogPanel>
            </Transition.Child>
          </div>
        </div>
      </Dialog>
    </Transition>
  )
}

'use client'

import { useState } from 'react'

export default function JoinPublicGroupModal({ open, onClose }: { open: boolean; onClose: () => void }) {
  const [selectedGroup, setSelectedGroup] = useState('')

  const availableGroups = ['Math Club', 'History Crashers', 'JavaScript Ninjas']

  if (!open) return null

  return (
    <div className="fixed inset-0 bg-black/30 flex items-center justify-center z-50">
      <div className="bg-white p-6 rounded-xl shadow-lg w-full max-w-md">
        <h2 className="text-xl font-semibold mb-4">Join Public Group</h2>
        <select
          className="w-full px-4 py-2 border rounded mb-4"
          value={selectedGroup}
          onChange={e => setSelectedGroup(e.target.value)}
        >
          <option value="">Select a group</option>
          {availableGroups.map(group => (
            <option key={group} value={group}>{group}</option>
          ))}
        </select>
        <div className="flex justify-end gap-2">
          <button onClick={onClose} className="px-4 py-2 text-gray-600">Cancel</button>
          <button
            onClick={() => {
              console.log('Joining:', selectedGroup)
              onClose()
            }}
            className="px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700"
          >
            Join
          </button>
        </div>
      </div>
    </div>
  )
}

'use client'

import { createContext, useContext, useState, useEffect, ReactNode } from 'react'
import axios from 'axios'

interface Group {
  id: string
  name: string
  type: 'public' | 'private'
  description?: string 
}

interface GroupContextType {
  group: Group | null
  setGroup: (group: Group | null) => void 
}

const GroupContext = createContext<GroupContextType | undefined>(undefined)

export const GroupProvider = ({ groupId, children, initialGroup }: { groupId: string, children: ReactNode, initialGroup?: Group }) => {
  const [group, setGroup] = useState<Group | null>(initialGroup || null)

  useEffect(() => {
    if (initialGroup) {
      setGroup(initialGroup);
      return;
    }

    const getGroupInfo = async () => {
      try {
        const response = await axios.get(`http://localhost:8080/groups/${groupId}`, {
          withCredentials: true
        })
        setGroup(response.data) 
      } catch (error) {
        console.error("Failed to fetch group info:", error)
        setGroup(null)
      }
    }
    
    getGroupInfo()
  }, [groupId, initialGroup])

  return (
    <GroupContext.Provider value={{ group, setGroup }}>
      {children}
    </GroupContext.Provider>
  )
}

export const useGroup = () => {
  const context = useContext(GroupContext)
  if (!context) {
    throw new Error('useGroup must be used within a GroupProvider')
  }
  return context
}

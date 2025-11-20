'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { useGroup } from './GroupContext'
import { 
  MessageCircle, 
  ListTodo, 
  UserPlus, 
  Settings,
  Users,
  BarChart3 // Placeholder for logo
} from 'lucide-react'

// Skeleton Loader for when group isn't loaded yet
function SidebarSkeleton() {
  return (
    <nav className="w-64 bg-white p-6 border-r flex flex-col justify-between shadow-lg">
      <div>
        <div className="flex items-center gap-2 mb-10 px-2">
          <BarChart3 size={32} className="text-purple-600" />
          <div className="h-6 w-32 bg-gray-200 rounded-md animate-pulse"></div>
        </div>
        <ul className="space-y-3">
          {[...Array(4)].map((_, i) => (
            <li key={i} className="h-10 bg-gray-200 rounded-md animate-pulse"></li>
          ))}
        </ul>
      </div>
    </nav>
  )
}

export default function GroupSidebar({ groupId }: { groupId: string }) {
  const { group } = useGroup()
  const pathname = usePathname()

  if (!group) {
    return <SidebarSkeleton />
  }

  const navLinks = [
    { href: `/groups/${groupId}/chat`, label: 'Chat', icon: MessageCircle },
    { href: `/groups/${groupId}/tasks`, label: 'Tasks', icon: ListTodo },
    { href: `/groups/${groupId}/members`, label: 'Members', icon: Users },
    // Conditionally include 'Join Requests' only if the group is private
    ...(group.type === 'private' 
      ? [{ href: `/groups/${groupId}/join-requests`, label: 'Join Requests', icon: UserPlus }] 
      : []),
    { href: `/groups/${groupId}/settings`, label: 'Settings', icon: Settings },
  ]

  return (
    <nav className="w-64 bg-white p-6 border-r flex-col justify-between shadow-lg hidden md:flex">
      {/* Group Name Header */}
      <div className="flex-grow">
        <div className="flex items-center gap-3 mb-10 px-2">
          {/* Avatar Placeholder */}
          <div className="flex-shrink-0 w-10 h-10 rounded-full bg-purple-100 flex items-center justify-center border-2 border-purple-200">
            <span className="text-lg font-semibold text-emerald-400">
              {group.name ? group.name[0].toUpperCase() : <Users size={20} />}
            </span>
          </div>
          <h2 className="text-xl font-bold text-gray-800 truncate" title={group.name}>
            {group.name}
          </h2>
        </div>

        {/* Navigation Links */}
        <ul className="space-y-3">
          {navLinks.map(({ href, label, icon: Icon }) => {
            const isActive = pathname === href;
            return (
              <li key={href}>
                <Link
                  href={href}
                  className={`
                    flex items-center gap-3 px-3 py-3 rounded-lg font-medium transition-colors
                    ${
                      isActive 
                        ? 'bg-amber-400 text-white shadow-md' 
                        : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900'
                    }
                  `}
                >
                  <Icon size={20} />
                  <span>{label}</span>
                </Link>
              </li>
            );
          })}
        </ul>
      </div>
      
      {/* Footer can be added here if needed */}
    </nav>
  )
}

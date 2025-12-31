import { notFound } from 'next/navigation'
import { cookies } from 'next/headers'
import GroupSidebar from './GroupSidebar'
import { GroupProvider } from './GroupContext'
import JoinLayout from './JoinLayout'
import axios from 'axios'
import { Clock } from 'lucide-react' // Added icon

const fetchMembershipStatus = async (groupId: string, token: string | undefined) => {
  try {
    const res = await axios.get(`http://localhost:8080/groups/${groupId}/user/status`, {
      headers: {
        Cookie: `token=${token}`
      }
    })
    return res.data
  } catch(err) {
    console.error(err)
    return null
  }
}

export default async function GroupLayout({ children, params }) {
  const { groupId } = params
  const cookieStore = cookies()
  const token = cookieStore.get('token')?.value

  const data = await fetchMembershipStatus(groupId, token)

  if (!data || !data.status) {
    notFound()
  }

  const { status, group } = data; 

  if (status === "not_a_member") {
    return <JoinLayout groupId={groupId} token={token} groupInfo={group} />
  }

  if (status === "pending") {
    return (
      <div className="flex min-h-screen justify-center items-center bg-gray-50 p-4">
        <div className="bg-white p-8 rounded-xl shadow-lg w-full max-w-md text-center">
          <div className="mx-auto flex items-center justify-center h-16 w-16 rounded-full bg-amber-100 mb-6">
            <Clock className="h-8 w-8 text-amber-600" />
          </div>
          <h2 className="text-2xl font-semibold text-gray-800 mb-3">Join Request Pending</h2>
          <p className="text-gray-600">
            Your request to join <span className="font-semibold">{group?.name || 'this group'}</span> is pending approval.
          </p>
          <p className="text-gray-600 mt-2">Please wait for an admin to accept your request.</p>
        </div>
      </div>
    )
  }

  return (
    <GroupProvider groupId={groupId} initialGroup={group}>
      <div className="flex min-h-screen">
        <GroupSidebar groupId={groupId} />
        <main className="flex-1 p-6 md:p-10 bg-gray-50 overflow-y-auto">
          {children}
        </main>
      </div>
    </GroupProvider>
  )
}
